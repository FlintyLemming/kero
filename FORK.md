# 个人 fork 说明

这是 [egoist/kero](https://github.com/egoist/kero) 的个人 fork，不向上游提 PR。
这里维护的是 **v1**，也就是 Swift/AppKit 版本，因为 2.x（Go 重写）去掉了 Alacritty 后端。

## 分支

- `v1-custom`：本 fork 的默认分支，基于上游 `v1`（0.1.48），自己的改动都在这里。
- `v1`：上游 `v1` 的镜像。上游已经不再更新它，所以基本不用同步；
  万一上游又修了 v1，用 `git fetch upstream && git switch v1-custom && git rebase upstream/v1`。
- `main`：上游 2.x 的镜像。`custom`：之前在 2.x 上做的改动，留作存档。

远程：`upstream` 指向 egoist/kero（推送地址设为 `DISABLE`，推不上去），`origin` 指向本 fork。

## 构建

需要 Xcode、Rust（Alacritty 后端的桥接库）和 Bun。第一次构建前先拉取子模块：

```bash
git submodule update --init --recursive
xcodebuild -project kero.xcodeproj -scheme kero -configuration Debug \
  -destination 'platform=macOS,arch=arm64' -scmProvider system \
  CODE_SIGNING_ALLOWED=NO CODE_SIGNING_REQUIRED=NO build
```

`-scmProvider system` 让 Swift 包用系统的 git 拉取。Xcode 自带的 git 会去读钥匙串里 `github.com` 的凭据，
弹出钥匙串密码框；这些包都是公开的，那个框直接拒绝就行。

命令行构建不会把新字符串写进 `kero/Localizable.xcstrings`，加了界面文字要手动补上中文和日文翻译。

Debug 构建的包名是 `sh.kero.dev`，设置在 `~/.config/kero-dev/`，可以和装好的 Kero 同时运行，互不影响。

## 打包

```bash
scripts/fork-package.sh
```

输出到 `build/fork/`：`Kero.app`、`Kero-<版本>.dmg` 和 `.zip`。

- 只构建 Apple Silicon（arm64）版本，使用 ad-hoc 签名，不需要苹果开发者账号。
  Apple Silicon 不运行完全没签名的程序，ad-hoc 签名就是最低要求，它不对应任何开发者身份。
- 关闭了 hardened runtime：它只在公证时才有用，而且开着时 ad-hoc 签名的应用可能加载不了自己内置的 ad-hoc 框架。
- 版本号是 `<上游版本>-patch-<yymmdd>`，比如 `0.1.48-patch-261113`；build 号是 `<yymmdd>.<HHMM>`，只会往上涨，
  Sparkle 靠它判断新旧。这两个值都是构建时传给 xcodebuild 的，不提交进仓库。
- 包名是 `sh.kero`，和官方版一样，所以替换 /Applications 里的 Kero 后，设置和会话都会照旧保留。
- 在本机构建的应用没有隔离属性（quarantine），可以直接打开。从网上下载的 dmg 或 zip 会被 Gatekeeper 拦下，
  装好后运行一次 `xattr -dr com.apple.quarantine /Applications/Kero.app`。
- `DERIVED_DATA=<目录>` 可以复用已有的构建目录，默认用 `build/DerivedData`。

## 更新

Sparkle 的更新源（`kero/Info.plist` 里的 `SUFeedURL`）指向本仓库的
`https://github.com/FlintyLemming/kero/releases/latest/download/appcast.xml`，`SUPublicEDKey` 是本 fork 自己的公钥，
所以上游往 releases.kero.sh 发的任何版本都不会覆盖这个构建。这个仓库发出第一个带 appcast 的 Release 之前，
"检查更新"会报错，可以不用管。

签名私钥在 `~/Library/Application Support/kero-fork/sparkle_ed25519.key`，没有放进钥匙串，也不在仓库里。
需要时用 Sparkle 的 `sign_update --ed-key-file` 或 `generate_appcast --ed-key-file` 读取。
已安装的构建只接受用这把私钥签名的更新，**要备份到密码管理器**。

## 本 fork 的功能

### 项目默认命令

在侧边栏右键项目 → **默认命令…**，填写一行命令（如 `claude` 或 `npm run dev`），留空则清除。

- 在这个项目里新开的标签页（+ 按钮、⌘T、命令面板）会先在交互式登录 shell 里运行这条命令，
  命令结束后回到同一个进程里的登录 shell。命令按你自己 shell 的语法写，alias 也能用。
- 分割出来的窗格、重启应用后恢复的会话、`kero` 命令行和自动化接口打开的终端都不受影响，直接进入 shell。
- 命令跟项目一起保存在会话快照里，重启应用后还在。

代码：`kero/ProjectDefaultCommandPrompt.swift`（对话框），`Project.defaultCommand`，
`TerminalSession` 启动脚本里的 `startupCommand`，以及 `SessionStore`/`TerminalManager` 里的保存和恢复。
