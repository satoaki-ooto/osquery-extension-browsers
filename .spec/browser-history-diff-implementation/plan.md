# ブラウザ履歴差分検出機能 - 実装方針

## 概要

osquery extensionでブラウザ履歴を取得する際に、remote setting経由での定期実行で差分を検出できるようにする機能の実装方針。

**作成日**: 2025-01-19  
**対象**: osquery-extension-browsers

---

## 背景と課題

### 現在の実装の問題点

[`internal/browsers/chromium/history.go`](../internal/browsers/chromium/history.go)の[`FindHistory()`](../internal/browsers/chromium/history.go:38)関数は：

1. **全履歴を毎回返す**: `ORDER BY last_visit_time DESC`で全履歴エントリを取得
2. **差分検出の仕組みがない**: 前回実行時からの変更を追跡する機能がない
3. **状態管理がない**: 最後に取得した履歴の時刻を記録していない
4. **Remote settingとの相性が悪い**: 定期実行しても常に全データを返すため差分が見えない

### ユーザー要件

- Remote setting経由で数秒に1度の高頻度実行
- Forensic調査（osqueryi）での柔軟なクエリ実行も必要
- WHERE句をcode側で処理すると汎用性が失われる

---

## 採用する設計: ハイブリッドアプローチ

### コンセプト

2つのテーブルを提供し、用途に応じて使い分ける：

| テーブル名 | 用途 | 動作 |
|-----------|------|------|
| `browser_history` | Forensic調査、アドホック分析 | 全履歴を返す（既存の動作） |
| `browser_history_mon` | Remote setting、リアルタイム監視 | 前回実行時からの差分のみを返す |

### メリット

1. ✅ **汎用性**: 用途に応じてテーブルを選択可能
2. ✅ **パフォーマンス**: 差分テーブルは高頻度実行でも効率的
3. ✅ **柔軟性**: WHERE句を使わずに差分を実現
4. ✅ **後方互換性**: 既存の`browser_history`テーブルはそのまま維持
5. ✅ **シンプル**: 状態管理は軽量なJSONファイル

---

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────┐
│                         osquery                              │
├─────────────────────────────────────────────────────────────┤
│  Remote Setting          │         osqueryi                  │
│  (監視用)                │      (Forensic調査用)             │
│                          │                                   │
│  SELECT * FROM           │  SELECT * FROM browser_history    │
│  browser_history_mon    │  WHERE url LIKE '%example.com%'   │
└────────┬─────────────────┴──────────────┬───────────────────┘
         │                                 │
         │                                 │
┌────────▼─────────────────────────────────▼───────────────────┐
│              osquery Extension                                │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌─────────────────────┐    ┌──────────────────────┐        │
│  │ browser_history     │    │ browser_history_mon │        │
│  │ (全履歴)            │    │ (差分のみ)           │        │
│  └──────────┬──────────┘    └──────────┬───────────┘        │
│             │                           │                     │
│             │                           │                     │
│             │                  ┌────────▼────────┐           │
│             │                  │  State Manager  │           │
│             │                  │ (状態管理)      │           │
│             │                  └────────┬────────┘           │
│             │                           │                     │
│             │                           ▼                     │
│             │              /tmp/browser_history_state.json    │
│             │                                                 │
│  ┌──────────▼─────────────────────────────────────────────┐ │
│  │         FindHistory() with Options                      │ │
│  │  - MinTime: 時刻範囲フィルタ                            │ │
│  │  - MaxTime: 時刻範囲フィルタ                            │ │
│  │  - Limit: 件数制限                                      │ │
│  └──────────┬──────────────────────────────────────────────┘ │
│             │                                                 │
└─────────────┼─────────────────────────────────────────────────┘
              │
              │
┌─────────────▼─────────────────────────────────────────────────┐
│                    Browser Databases                           │
├────────────────────────────────────────────────────────────────┤
│  Chrome History DB  │  Firefox History DB  │  Edge History DB  │
└────────────────────────────────────────────────────────────────┘
```

---

## 実装詳細

### 1. テーブルスキーマ

#### browser_history（既存）

```go
columns := []table.ColumnDefinition{
    table.TextColumn("time"),           // 訪問時刻（人間可読形式）
    table.TextColumn("title"),          // ページタイトル
    table.TextColumn("url"),            // URL
    table.TextColumn("profile"),        // プロファイルID
    table.TextColumn("browser_type"),   // ブラウザタイプ
    table.IntegerColumn("unix_time"),   // Unix timestamp（追加）
    table.TextColumn("change_type"),    // "new" or "updated"（メタデータ）
}
```

#### browser_history_mon（新規）

```go
columns := []table.ColumnDefinition{
    table.TextColumn("time"),
    table.TextColumn("title"),
    table.TextColumn("url"),
    table.TextColumn("profile"),
    table.TextColumn("browser_type"),
    table.IntegerColumn("unix_time"),
    table.TextColumn("change_type"),    // "new" or "updated"（メタデータ）
}
```

**注**: 両テーブルは同じスキーマを持ち、違いは取得するデータの範囲のみです。
- `browser_history`: 全履歴を返す
- `browser_history_mon`: 前回実行時からの差分のみを返す
- `visit_count`は内部的に使用しますが、テーブルには含めません（`change_type`の判定に使用）

### 2. 状態管理

#### ファイル構造

```json
{
  "last_fetch_time": {
    "chrome_Default": "2025-01-19T10:00:00Z",
    "chrome_Profile 1": "2025-01-19T10:00:00Z",
    "firefox_default-release": "2025-01-19T09:55:00Z"
  }
}
```

#### 保存場所

- **デフォルト**: `/tmp/browser_history_state.json`
- **理由**: 
  - 軽量なデータ（数KB程度）
  - 再起動時にリセットされても問題ない（初回実行として扱える）
  - 書き込み権限の問題が少ない

#### 並行制御

- `sync.RWMutex`でロック制御
- 読み込み時: RLock（複数goroutineから同時読み込み可能）
- 書き込み時: Lock（排他制御）

### 3. FindHistory()の拡張

#### 新しいシグネチャ

```go
type FindHistoryOptions struct {
    MinTime time.Time  // この時刻以降のみ取得（ゼロ値の場合は全件）
    MaxTime time.Time  // この時刻以前のみ取得（ゼロ値の場合は制限なし）
    Limit   int        // 取得件数制限（0の場合は制限なし）
}

func FindHistory(profile common.Profile, opts *FindHistoryOptions) ([]common.HistoryEntry, error)
```

#### SQLクエリの動的生成

```sql
SELECT id, url, title, last_visit_time, visit_count
FROM urls
WHERE 1=1
  AND last_visit_time > ?  -- MinTimeが指定された場合
  AND last_visit_time < ?  -- MaxTimeが指定された場合
ORDER BY last_visit_time DESC
LIMIT ?                    -- Limitが指定された場合
```

### 4. 差分判定ロジック

```go
func determineChangeType(entry common.HistoryEntry, lastFetchTime time.Time) string {
    if lastFetchTime.IsZero() {
        return "new"  // 初回実行
    }
    
    if entry.VisitCount == 1 {
        return "new"  // 新規訪問
    }
    
    return "updated"  // 既存URLへの再訪問
}
```

---

## 使用例

### Remote Setting（監視用）

```json
{
  "schedule": {
    "browser_history_monitoring": {
      "query": "SELECT * FROM browser_history_mon",
      "interval": 5,
      "description": "5秒ごとに新しいブラウザ履歴を監視"
    }
  }
}
```

**実行結果例**:

```
time                | url                    | change_type | profile
--------------------|------------------------|-------------|----------
2025-01-19 10:05:23 | https://example.com    | new         | Default
2025-01-19 10:05:25 | https://github.com     | updated     | Default
```

### Forensic調査（osqueryi）

```sql
-- 全履歴から特定URLを検索
SELECT * FROM browser_history
WHERE url LIKE '%example.com%'
ORDER BY time DESC;

-- 特定期間の履歴を取得
SELECT * FROM browser_history
WHERE unix_time BETWEEN 1705651200 AND 1705737600;

-- 新規訪問のみを分析
SELECT * FROM browser_history
WHERE change_type = 'new'
ORDER BY time DESC
LIMIT 100;

-- 再訪問（頻繁にアクセスしているサイト）を分析
SELECT * FROM browser_history
WHERE change_type = 'updated'
ORDER BY time DESC;
```

---

## 実装ファイル

### 新規作成

1. **`internal/browsers/common/state.go`**
   - 状態管理の実装
   - JSONファイルの読み書き
   - 並行制御

### 変更

1. **`internal/browsers/chromium/history.go`**
   - `FindHistoryOptions`構造体の追加
   - `FindHistory()`のシグネチャ変更
   - SQLクエリの動的生成

2. **`internal/browsers/firefox/history.go`**
   - Chromiumと同様の変更

3. **`cmd/browser_extend_extension/main.go`**
   - `browserHistoryMonTablePlugin()`の追加
   - `generateBrowserHistoryMon()`の実装
   - 既存の`generateBrowserHistory()`の更新（`change_type`追加）

4. **`internal/browsers/common/interfaces.go`**
   - `FindHistory()`のインターフェース更新

---

## パフォーマンス考慮事項

### 初回実行時の対策

初回実行時は全履歴が返されるため、大量のデータが返る可能性がある。

**対策**:
- `Limit`オプションで件数制限（例: 最新1000件のみ）
- Remote settingの`interval`を長めに設定（初回のみ）

### 状態ファイルのサイズ

- プロファイル数 × 8バイト（timestamp） ≈ 数百バイト
- 100プロファイルでも1KB未満

### SQLクエリのパフォーマンス

- `last_visit_time`カラムにはインデックスが存在（Chromeのデフォルト）
- 時刻範囲フィルタは効率的に実行される

---

## エラーハンドリング

### 状態ファイルの読み込み失敗

```go
if err := state.Load(); err != nil {
    log.Printf("Failed to load state, treating as first run: %v", err)
    // 初回実行として扱う（エラーにしない）
}
```

### ブラウザDBへのアクセス失敗

```go
if err != nil {
    log.Printf("Failed to find history for profile %s: %v", profile.ID, err)
    continue  // 他のプロファイルの処理を継続
}
```

### 状態ファイルの保存失敗

```go
if err := state.Save(); err != nil {
    log.Printf("Failed to save state: %v", err)
    // 次回実行時に再度全件取得される（データ損失はない）
}
```

---

## テスト戦略

### ユニットテスト

1. **状態管理**
   - `state.Load()` / `state.Save()`
   - 並行アクセス時の動作

2. **FindHistory()**
   - オプションなし（全件取得）
   - MinTime指定
   - MaxTime指定
   - Limit指定

3. **差分判定**
   - 初回実行時
   - 新規訪問
   - 再訪問

### 統合テスト

1. **差分テーブルの動作**
   - 初回実行 → 全件返却
   - 2回目実行 → 差分のみ返却
   - 状態ファイルの永続化

2. **複数ブラウザ/プロファイル**
   - Chrome + Firefox
   - 複数プロファイル

---

## マイグレーション計画

### Phase 1: 基盤実装

1. `internal/browsers/common/state.go`の実装
2. `FindHistoryOptions`の追加
3. `FindHistory()`の拡張

### Phase 2: 差分テーブル実装

1. `browser_history_mon`テーブルの追加
2. `generateBrowserHistoryDiff()`の実装
3. 状態管理の統合

### Phase 3: テストと検証

1. ユニットテストの作成
2. 統合テストの実行
3. パフォーマンステスト

### Phase 4: ドキュメント更新

1. README.mdの更新
2. 使用例の追加
3. トラブルシューティングガイド

---

## 代替案との比較

### 代替案1: WHERE句ベースのフィルタリング

**概要**: osqueryのWHERE句で時刻範囲を指定

**却下理由**:
- Remote settingで数秒に1度の実行では、WHERE句の時刻調整が困難
- Forensic用途での柔軟性が失われる
- Code側でWHERE句を処理すると汎用性が低下

### 代替案2: 単一テーブル + モードフラグ

**概要**: `browser_history`テーブルに`mode`カラムを追加

**却下理由**:
- テーブルスキーマが複雑化
- 用途の違いが明確でない
- 既存のクエリに影響を与える可能性

---

## 今後の拡張可能性

### 1. 差分の保持期間設定

```go
type StateConfig struct {
    RetentionDays int  // 古い状態を自動削除
}
```

### 2. 状態ファイルの保存場所設定

```go
// 環境変数で設定可能に
stateFilePath := os.Getenv("BROWSER_HISTORY_STATE_PATH")
if stateFilePath == "" {
    stateFilePath = "/tmp/browser_history_state.json"
}
```

### 3. 差分の詳細情報

```go
type HistoryDiff struct {
    Entry      HistoryEntry
    ChangeType string  // "new", "updated", "deleted"
    OldValue   *HistoryEntry  // 更新前の値
}
```

---

## セキュリティ考慮事項

### 状態ファイルのアクセス権限

- ファイルパーミッション: `0644`（所有者のみ書き込み可能）
- 機密情報は含まれない（タイムスタンプのみ）

### ブラウザDBへのアクセス

- 読み取り専用モード: `mode=ro&immutable=1`
- ロックを取得しない（ブラウザの動作に影響しない）

---

## まとめ

このハイブリッドアプローチにより：

1. ✅ Remote settingでの高頻度監視が効率的に実現
2. ✅ Forensic調査での柔軟なクエリ実行が可能
3. ✅ 既存機能への影響なし（後方互換性）
4. ✅ シンプルな実装（状態管理は軽量なJSON）
5. ✅ 拡張性が高い（将来的な機能追加が容易）

---

## 参考資料

- [osquery Table Plugin API](https://osquery.readthedocs.io/en/stable/development/osquery-sdk/)
- [Chrome History Database Schema](https://forensics.wiki/google_chrome/#history)
- [Firefox Places Database](https://forensics.wiki/mozilla_firefox/#places-database)
