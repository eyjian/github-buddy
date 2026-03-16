入门openspec

OpenSpec是一个AI友好的规范驱动开发工具，官网定义为 "A lightweight framework spec-driven development"（轻量级规范驱动框架）。

# 官方文档

[https://github.com/Fission-AI/OpenSpec](https://github.com/Fission-AI/OpenSpec)

# 安装openspec

```shell
npm install -g @fission-ai/openspec@latest
```

安装依赖 Node.js 20.19.0 或以上版本。

# 配置openspec

## 初始化openspec

进入到工作区（Workspace）目录（一个工作区可包含多个项目）后执行：

```shell
openspec init
```

执行后，会出现如下选择提示：

* /opsx:new      Create a change
* /opsx:continue Next artifact
* /opsx:appy     Implement tasks

这里会提示选择一个工具（tools），支持丰富的工具，比如：

* Claude Code
* Gemini CLI
* Cursor
* Opencode
* Qwen code
* Trae
* CodeBuddy Code (CLI)

注意按空格键选中后再按回车键确认。如果需要更改，再次执行“openspec init”即可。

假设工作区目录为 /data/workspace，工具选择“CodeBuddy Code (CLI) ”，完成后可看到如下：

```
[/data/workspace]# tree /data/workspace/.codebuddy/commands/
/data/workspace/.codebuddy/commands/
└── opsx
    ├── apply.md
    ├── archive.md
    ├── explore.md
    └── propose.md

1 directory, 4 files

[/data/workspace]# tree /data/workspace/.codebuddy/skills/
/data/workspace/.codebuddy/skills/
├── openspec-apply-change
│   └── SKILL.md
├── openspec-archive-change
│   └── SKILL.md
├── openspec-explore
│   └── SKILL.md
└── openspec-propose
    └── SKILL.md

4 directories, 4 files
```

可看到安装了command和skill。这是openspec-1.2.0的效果，执行命令“openspec --version”可查看openspec的版本号。

开启“/opsx:new, /opsx:continue, /opsx:ff, /opsx:verify, /opsx:sync, /opsx:bulk-archive, /opsx:onboard”等扩展command方法，在工作区目录下执行命令“openspec config profile”：

```
✔ What do you want to configure? Delivery and workflows
✔ Delivery mode (how workflows are installed): Both (skills + commands) [current]
? Select workflows to make available:
❯[x] Propose change
 [x] Explore ideas
 [ ] New change
 [ ] Continue change
 [x] Apply tasks
 [ ] Fast-forward
 [ ] Sync specs
 [x] Archive change
 [ ] Bulk archive
 [ ] Verify change
 [ ] Onboard

Create proposal, design, and tasks from a request
Space to toggle, Enter to confirm
```

“[x]”为当前已经安装好的command和skill，根据需要按空格添加新的或取消已有的，然后按回车键确认。

| 命令 | 功能说明 |
| :--- | :--- |
| `/opsx:propose` | 创建变更并一步生成规划制品（默认快捷流程） |
| `/opsx:explore` | 梳理思路、调研问题、明确需求 |
| `/opsx:new` | 新建变更脚手架（扩展工作流） |
| `/opsx:continue` | 创建下一个制品（扩展工作流） |
| `/opsx:ff` | 快进生成规划制品（扩展工作流） |
| `/opsx:apply` | 执行任务实现，并按需更新制品 |
| `/opsx:verify` | 对照制品校验实现结果（扩展工作流） |
| `/opsx:sync` | 将增量规范同步到主干（扩展工作流，可选） |
| `/opsx:archive` | 完成后归档 |
| `/opsx:bulk-archive` | 批量归档多个已完成变更（扩展工作流） |
| `/opsx:onboard` | 引导式完成一次完整端到端变更（扩展工作流） |

# “/opsx:propose”和“/opsx:new”

执行“/opsx:propose”或者“/opsx:new”开始一个新需求。

“/opsx:propose”和“/opsx:new”的区别：

* /opsx:propose 一键搞定：提出变更 + 直接生成完整的规划制品（需求、步骤、验收标准），走「快捷流程」；
* /opsx:new 只搭架子：仅创建一个空的变更脚手架（无实际内容），是「扩展流程」的第一步，后续需要手动补内容。

## 详细区别：

| 维度         | /opsx:propose（提议）                          | /opsx:new（新建）                              |
|--------------|------------------------------------------------|------------------------------------------------|
| 核心动作     | 「提出变更想法」+「生成完整规划」一步到位       | 仅创建一个空的变更框架（脚手架），无任何规划内容 |
| 流程属性     | 默认「快捷流程」，适合简单变更                 | 「扩展流程」的起点，适合复杂变更               |
| 产出物       | 有实际内容的规划制品（需求、验收标准、实现步骤） | 空的变更目录 / 模板（只有结构，没有具体信息）  |
| 后续操作     | 可直接跳到 /opsx:apply 开始开发                 | 必须接着用 /opsx:continue 逐步补全规划内容     |
| 适用场景     | 小改动（比如改个文案、修复一个小 bug）          | 大功能 / 复杂变更（比如开发一个新模块、重构核心逻辑） |

## 举个实际例子（以github-buddy项目为例）

### 场景 1：用 /opsx:propose

想给 github-buddy 加一个「自动刷新IP的小开关」（简单变更），执行“/opsx:propose”→工具直接生成：

* 变更目标：新增 IP 自动刷新开关
* 验收标准：开关打开后每小时自动更新 IP
* 实现步骤：1. 加配置项 2. 写定时逻辑 3. 测试
（一步到位，直接能开始写代码）

### 场景 2：用 /opsx:new

想给github-buddy重构整个IP检测逻辑（复杂变更），执行“/opsx:new”→工具只生成一个变更框架：

* 变更名称：[待填写]
* 变更目标：[待填写]
* 验收标准：[待填写]
（只有结构，需要接着用“/opsx:continue”一步步补全需求、设计、验收标准等内容）

## 总结

* 简单变更（小修小改）：直接用“/opsx:propose”，省时间，一键到位；
* 复杂变更（大功能 / 重构）：先用“/opsx:new”搭架子，再用“/opsx:continue”分步完善，流程更可控；
* 核心差异：propose是「提想法 + 做规划」，new只是「建空架子」。

## 使用openspec

执行“/opsx:propose”开始一个新需求。提示词：根据文件ai_requirements.md的内容执行，放在/data/workspace/github.com/eyjian/github-buddy目录下。完成后可到openspec/changes/github-buddy目录下查看生成的文件。

```
[/data/workspace/github.com/eyjian/github-buddy]# tree openspec/
openspec/
└── changes
    └── github-buddy
        ├── design.md
        ├── proposal.md
        ├── specs
        │   ├── auto-update
        │   │   └── spec.md
        │   ├── backup-rollback
        │   │   └── spec.md
        │   ├── cli-commands
        │   │   └── spec.md
        │   ├── hosts-management
        │   │   └── spec.md
        │   └── ip-detection
        │       └── spec.md
        └── tasks.md

8 directories, 8 files
```

**各目录和文件说明：**

| 路径 | 类型 | 说明 |
|------|------|------|
| openspec/ | 根目录 | OpenSpec 规范驱动开发的核心目录，存放所有与项目变更相关的规划、设计、规范和任务制品 |
| openspec/changes/ | 目录 | 项目所有变更（change）的根目录，每个子目录对应一个独立的变更项（此处仅包含 `github-buddy` 一个核心变更） |
| openspec/changes/github-buddy/ | 目录 | 「GitHub-Buddy 工具开发」变更项的专属目录，聚合该变更的全量 OpenSpec 制品 |
| openspec/changes/github-buddy/design.md | 文件 | 技术设计 — 采用 C++17 + CMake 构建，使用 CLI11/spdlog/libcurl 等库，平台抽象层设计，关键决策与风险分析 |
| openspec/changes/github-buddy/proposal.md | 文件 | 变更提议 — 阐述为何需要此工具（解决 GitHub 网络访问痛点），以及变更内容（5 个核心 capability） |
| openspec/changes/github-buddy/specs/ | 目录 | 5 个 spec 文件，详细需求规范 — 覆盖 IP 检测、Hosts 管理、备份回滚、自动更新、CLI 命令 5 个模块 |
| openspec/changes/github-buddy/specs/auto-update/ | 目录 | 「IP 自动更新」功能模块的专属目录 |
| openspec/changes/github-buddy/specs/auto-update/spec.md | 文件 | 定时自动更新（缓存机制、过期检查、配置管理） |
| openspec/changes/github-buddy/specs/backup-rollback/ | 目录 | 「hosts 备份与回滚」功能模块的专属目录 |
| openspec/changes/github-buddy/specs/backup-rollback/spec.md | 文件 | hosts 安全保障（自动备份、一键回滚、完整性校验） |
| openspec/changes/github-buddy/specs/cli-commands/ | 目录 | 「CLI 命令」功能模块的专属目录 |
| openspec/changes/github-buddy/specs/cli-commands/spec.md | 文件 | 命令行界面（init/update/rollback/status + 错误处理 + 日志） |
| openspec/changes/github-buddy/specs/hosts-management/ | 目录 | 「hosts 文件管理」功能模块的专属目录 |
| openspec/changes/github-buddy/specs/hosts-management/spec.md | 文件 | 「跨平台 hosts 文件管理（自动识别、提权、标记区块、增量更新） |
| openspec/changes/github-buddy/specs/ip-detection/ | 目录 | 「IP 检测与筛选」功能模块的专属目录 |
| openspec/changes/github-buddy/specs/ip-detection/spec.md | 文件 | 智能 IP 检测与质量筛选（多维度检测、评分排序、兜底切换） |
| openspec/changes/github-buddy/tasks.md | 文件 | 实现任务清单 — 9 个任务分组、37 个可追踪的实现任务 |
