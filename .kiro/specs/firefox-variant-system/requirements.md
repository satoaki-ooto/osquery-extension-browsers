# Requirements Document

## Introduction

この機能は、FirefoxブラウザでChromeのvariant仕組みと同様のアーキテクチャを実装することを目的としています。現在、Chromiumベースのブラウザ（Chrome、Edge、Chromium）は統一されたvariant仕組みで管理されていますが、Firefoxは独立した実装になっています。この機能により、Firefoxも複数のvariant（Firefox、Firefox ESR、Firefox Developer Edition、Firefox Nightly等）を統一的に扱えるようになります。

## Requirements

### Requirement 1

**User Story:** 開発者として、Firefoxの複数のvariantを統一的に検出・管理したい。これにより、コードの一貫性を保ち、新しいFirefoxのvariantを簡単に追加できるようになる。

#### Acceptance Criteria

1. WHEN システムがFirefoxのvariantを検出する THEN システムは利用可能なすべてのFirefoxのvariant（Firefox、Firefox ESR、Firefox Developer Edition、Firefox Nightly）を返すSHALL
2. WHEN 各プラットフォーム（Windows、macOS、Linux）でvariantを検出する THEN システムは各プラットフォーム固有のパスとプロセス名を正しく識別するSHALL
3. WHEN Firefoxのvariantが存在しない THEN システムは空のリストを返し、エラーを発生させないSHALL

### Requirement 2

**User Story:** 開発者として、各Firefoxのvariantから履歴データを取得したい。これにより、ユーザーがどのFirefoxのvariantを使用していても、統一的に履歴データにアクセスできるようになる。

#### Acceptance Criteria

1. WHEN 特定のFirefoxのvariantから履歴を取得する THEN システムはそのvariantの履歴データベース（places.sqlite）から履歴エントリを取得するSHALL
2. WHEN 履歴エントリを返す THEN 各エントリにはvariant名（Firefox、Firefox ESR等）が含まれるSHALL
3. WHEN 履歴データベースが存在しないまたはアクセスできない THEN システムは適切なエラーメッセージを返すSHALL
4. WHEN 複数のFirefoxのvariantが存在する THEN システムは各variantから独立して履歴を取得できるSHALL

### Requirement 3

**User Story:** 開発者として、各Firefoxのvariantからプロファイル情報を取得したい。これにより、ユーザーがどのFirefoxのvariantを使用していても、統一的にプロファイル情報にアクセスできるようになる。

#### Acceptance Criteria

1. WHEN 特定のFirefoxのvariantからプロファイルを取得する THEN システムはそのvariantのprofiles.iniファイルからプロファイル情報を取得するSHALL
2. WHEN プロファイル情報を返す THEN 各プロファイルにはvariant名（Firefox、Firefox ESR等）が含まれるSHALL
3. WHEN profiles.iniファイルが存在しない THEN システムはデフォルトプロファイルを作成して返すSHALL
4. WHEN 複数のFirefoxのvariantが存在する THEN システムは各variantから独立してプロファイルを取得できるSHALL

### Requirement 4

**User Story:** 開発者として、Firefoxのvariant仕組みがChromeのvariant仕組みと同じインターフェースを持つようにしたい。これにより、コードの一貫性を保ち、メンテナンスを簡素化できる。

#### Acceptance Criteria

1. WHEN Firefoxのvariant構造体を定義する THEN ChromeのBrowserVariant構造体と同じフィールド（Name、Paths、Process）を持つSHALL
2. WHEN Firefoxのvariant検出関数を実装する THEN ChromeのDetectBrowserVariants()と同様のシグネチャを持つSHALL
3. WHEN 既存のFirefoxコードを更新する THEN 新しいvariant仕組みを使用するように変更されるSHALL
4. WHEN 新しいFirefoxのvariantを追加する THEN 最小限のコード変更で追加できるSHALL