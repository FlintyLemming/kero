# 个人 fork 说明

这是 [egoist/kero](https://github.com/egoist/kero) 的个人 fork，不向上游提 PR。

## 分支

- `main`：上游 `main` 的镜像，只做快进，不在上面提交。
- `custom`：本 fork 的默认分支，也就是自己的改动所在的分支。每个功能尽量压成一个提交，叠在上游代码之上。
- `v1`：上游留下的旧 Swift/AppKit 版本（0.1.48 起不再更新），只作存档。

远程：`upstream` 指向 egoist/kero（推送地址设为 `DISABLE`，推不上去），`origin` 指向本 fork。

## 同步上游

```bash
git fetch upstream && git switch main && git merge --ff-only upstream/main
git switch custom && git rebase main
git push --force-with-lease origin custom
```

已开启 `rerere`，同样的冲突解决过一次后会自动套用。

冲突处理约定：

- `locales/catalog_gen.go`、`fileicons_gen.go` 是生成的文件：取上游版本后运行 `go tool mygo generate` 重新生成，不手动合并。
- 不修改 `CHANGELOG.md` 和 `mygo.json` 里的版本号。
- 新逻辑放在新文件里，上游文件只留很少的挂钩调用。

## 本地构建

不启用 GitHub Actions：上游的发布 workflow 依赖作者的签名证书和 secrets。

```bash
go tool mygo dev                                         # 开发版 "Kero Dev"，数据在 kero-dev 目录，和正式版互不影响
go tool mygo build -sign - -skip-notarize -skip-dmg      # ad-hoc 签名的正式版 .app，输出到 build/
```

注意：`mygo.json` 的 `updates.github` 仍然指向 egoist/kero。自己构建的版本检查更新时会被官方版本覆盖，
长期使用前需要在一个单独的提交里改掉。

## 本 fork 的功能

### 项目默认命令

在侧边栏右键项目 → **默认命令…**，填写一行命令（如 `claude` 或 `npm run dev`），留空则清除。

- 在应用里新建的会话（+ 按钮、⌘T、命令面板、打开项目、工作树会话等）都会先在 shell 里运行这条命令，命令结束后回到 shell。
- 分割出来的窗格、快速终端和命令行的 `kero new` 不受影响，仍然直接进入 shell。
- 命令保存在 kerod 里，跟随项目。远程主机上的项目只有在那台主机也运行本 fork 构建的 kero-server 时才能使用，
  否则保存时会提示错误。

代码：`project_command.go`、`internal/daemon/project_command.go`，以及上游文件里的几处挂钩
（`proto.Project.Command`、`CreateParams.ProjectCommand`、`project.command` RPC、项目右键菜单）。
