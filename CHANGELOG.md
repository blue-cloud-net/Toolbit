# Changelog

本文件记录 Blue Cloud Toolbit 的所有显著变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

尚无变更。

## [0.1.0] - 2026-10-27

首个开发版本，仅包含工程地基与文档体系，工具尚未上线。

### Added

- 仓库骨架：`cmd/toolbit`、`internal/{api,catalog,codecsvc,cryptosvc,httpx,netguard,stats,toolerr,tools,webui}`、`web/`、`deploy/`、`catalog/`
- 文档体系：[README](README.md)、[AGENTS](AGENTS.md)、[架构总览](docs/architecture.md)、[开发规范](docs/standards/00-intro.md)、[API 契约](docs/api/README.md)、[工具规格](docs/tools.md)、[路线图](docs/roadmap.md)
- 工具目录 `catalog/tools.yaml`：登记 70 个工具的分类、传输方式、限流档位与限额
- CI 工作流骨架：`ci`（lint / test）、`builder`（构建 Tongsuo 基础镜像）、`docker`（多架构镜像推 GHCR）
- 文档校验：Markdown lint、相对链接检查、catalog / 错误码 / 环境变量覆盖率检查

[未发布]: https://github.com/blue-cloud-net/Toolbit/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/blue-cloud-net/Toolbit/releases/tag/v0.1.0
