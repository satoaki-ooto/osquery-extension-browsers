# Design Document

## Overview

この設計では、FirefoxブラウザにChromeのvariant仕組みと同様のアーキテクチャを実装します。現在のFirefoxの実装は単一のブラウザとして扱われていますが、実際にはFirefox、Firefox ESR、Firefox Developer Edition、Firefox Nightlyなど複数のvariantが存在します。

この設計により、以下の利点が得られます：
- コードの一貫性：ChromiumとFirefoxで同じアーキテクチャパターンを使用
- 拡張性：新しいFirefoxのvariantを簡単に追加可能
- 保守性：統一されたインターフェースによる保守の簡素化

## Architecture

### Current State
現在のFirefoxの実装：
```
firefox/
├── finder.go      - 単一のFirefoxパスを検出
├── history.go     - 履歴取得（variant情報なし）
└── profile.go     - プロファイル取得（variant情報なし）
```

### Target State
新しいFirefoxの実装：
```
firefox/
├── variants.go    - 複数のFirefoxのvariantを検出・管理
├── finder.go      - variant仕組みを使用するように更新
├── history.go     - variant情報を含む履歴取得
└── profile.go     - variant情報を含むプロファイル取得
```

## Components and Interfaces

### 1. BrowserVariant Structure
ChromiumのBrowserVariant構造体と同じインターフェースを持つFirefox用の構造体：

```go
type BrowserVariant struct {
    Name    string   // variant名（Firefox、Firefox ESR等）
    Paths   []string // データディレクトリのパス
    Process string   // プロセス名
}
```

### 2. DetectBrowserVariants Function
各プラットフォームでFirefoxのvariantを検出する関数：

```go
func DetectBrowserVariants() []BrowserVariant
```

### 3. Variant Detection Logic
各プラットフォームでの検出ロジック：

**Windows:**
- Firefox: `%APPDATA%\Mozilla\Firefox\Profiles`, プロセス: `firefox.exe`
- Firefox ESR: `%APPDATA%\Mozilla\Firefox\Profiles`, プロセス: `firefox.exe`
- Firefox Developer Edition: `%APPDATA%\Mozilla\Firefox\Profiles`, プロセス: `firefox.exe`
- Firefox Nightly: `%APPDATA%\Mozilla\Firefox\Profiles`, プロセス: `firefox.exe`

**macOS:**
- Firefox: `~/Library/Application Support/Firefox/Profiles`, プロセス: `Firefox`
- Firefox ESR: `~/Library/Application Support/Firefox/Profiles`, プロセス: `Firefox`
- Firefox Developer Edition: `~/Library/Application Support/Firefox/Profiles`, プロセス: `Firefox Developer Edition`
- Firefox Nightly: `~/Library/Application Support/Firefox/Profiles`, プロセス: `Firefox Nightly`

**Linux:**
- Firefox: `~/.mozilla/firefox`, プロセス: `firefox`
- Firefox ESR: `~/.mozilla/firefox`, プロセス: `firefox-esr`
- Firefox Developer Edition: `~/.mozilla/firefox`, プロセス: `firefox-developer-edition`
- Firefox Nightly: `~/.mozilla/firefox`, プロセス: `firefox-nightly`

## Data Models

### Updated Profile Structure
既存のcommon.Profile構造体を使用し、BrowserVariantフィールドに適切なvariant名を設定：

```go
type Profile struct {
    ID             string
    Name           string
    Path           string
    Email          string
    BrowserType    string // "Firefox"
    BrowserVariant string // "Firefox", "Firefox ESR", etc.
}
```

### Updated HistoryEntry Structure
既存のcommon.HistoryEntry構造体を使用し、BrowserVariantフィールドに適切なvariant名を設定：

```go
type HistoryEntry struct {
    ID             int64
    URL            string
    Title          string
    VisitTime      time.Time
    VisitCount     int
    ProfileID      string
    BrowserType    string // "Firefox"
    BrowserVariant string // "Firefox", "Firefox ESR", etc.
}
```

## Error Handling

### Variant Detection Errors
- ディレクトリが存在しない場合：空のリストを返す（エラーなし）
- アクセス権限がない場合：そのvariantをスキップして続行
- 予期しないエラー：ログに記録してそのvariantをスキップ

### History and Profile Errors
- データベースファイルが存在しない：適切なエラーメッセージを返す
- データベースが破損している：エラーメッセージを返す
- アクセス権限がない：エラーメッセージを返す

## Testing Strategy

### Unit Tests
1. **Variant Detection Tests**
   - 各プラットフォームでのvariant検出
   - 存在しないパスの処理
   - 複数のvariantが存在する場合の処理

2. **History Retrieval Tests**
   - 各variantからの履歴取得
   - variant情報の正確性
   - エラーケースの処理

3. **Profile Discovery Tests**
   - 各variantからのプロファイル取得
   - variant情報の正確性
   - profiles.iniが存在しない場合の処理

### Integration Tests
1. **End-to-End Tests**
   - 実際のFirefoxのvariantでのテスト
   - 複数のvariantが同時に存在する環境でのテスト

### Mock Tests
1. **File System Mocking**
   - 異なるプラットフォーム環境のシミュレーション
   - エラーケースのシミュレーション

## Implementation Phases

### Phase 1: Core Variant System
- `variants.go`ファイルの作成
- `BrowserVariant`構造体の定義
- `DetectBrowserVariants()`関数の実装

### Phase 2: Integration with Existing Code
- `finder.go`の更新（variant仕組みを使用）
- `history.go`の更新（variant情報を含む）
- `profile.go`の更新（variant情報を含む）

### Phase 3: Testing and Validation
- 単体テストの実装
- 統合テストの実装
- 既存機能の回帰テスト