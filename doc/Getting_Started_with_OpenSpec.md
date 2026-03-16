实战openspec

# 什么是OpenSpec？

OpenSpec 是一个 AI 友好的规范驱动开发工具，官网将其定义为 "A lightweight framework spec-driven development"（轻量级规范驱动框架）。

现在的 AI 世界有个大问题：每个 AI 都是一座孤岛，互相听不懂、连不上。本质上，OpenSpec 想解决的是 AI Agent 时代的“互操作性”问题 —— **让 AI 世界不再碎片化，而是像互联网一样互联互通。它不仅仅是一个文件格式，更是 AI 网络的“通用接口标准”**。

**核心机制：先定规矩，再干活**

OpenSpec 提倡开发者先写好一份标准化的“契约”（Spec），明确规定好输入和输出是什么。

* **对 AI Agent 而言：** 智能体只要读懂这份契约，不需要看底层代码就能直接调用工具。这就像乐高积木，只要接口标准统一，任何形状的积木都能拼在一起。
* **对整个生态而言：** 它建立了一套 AI 界的“普通话”。无论背后是哪个大模型，无论运行在什么硬件上，只要套上 OpenSpec 的标准外壳，任何 AI Agent 都能无障碍地理解、连接和使用。

**终极愿景：AI 能力的自由流动**

OpenSpec 试图消除不同模型、不同硬件、不同开发者之间的隔阂。它让 AI 能力像水和电一样，通过标准化的管道，在整个网络中自由流动、被任意 Agent 即插即用，从而构建一个真正协作的 AI 价值网络。


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

可看到安装了 command 和 skill。这是 openspec-1.2.0 的效果，执行命令“openspec --version”可查看 openspec 的版本号。

开启“/opsx:new, /opsx:continue, /opsx:ff, /opsx:verify, /opsx:sync, /opsx:bulk-archive, /opsx:onboard”等扩展 command 方法，在工作区目录下执行命令“openspec config profile”：

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

“[x]”为当前已经安装好的 command 和 skill，根据需要按空格添加新的或取消已有的，然后按回车键确认。

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

**制品**

在软件开发过程中产生的任何文件、数据、代码或实体都是制品，中间代码、测试脚本、甚至本地运行的临时环境，这些都是制品 (Artifacts)。所有的交付件，本质上都是开发过程中产生的制品；但开发过程中产生的大量制品（如测试日志、中间代码），并不会作为交付件交给客户。总之只要是过程中"产出的东西"，都是（包括中间产物）制品，而制品中那些被合同/协议约定要交给对方的部分为交付件。

# “/opsx:propose”和“/opsx:new”

执行“/opsx:propose”或者“/opsx:new”开始一个新需求。“/opsx:propose”和“/opsx:new”的区别：

* **/opsx:propose** 一键搞定：提出变更 + 直接生成完整的规划制品（需求、步骤、验收标准），走「快捷流程」；
* **/opsx:new** 只搭架子：仅创建一个空的变更脚手架（无实际内容），是「扩展流程」的第一步，后续需要手动补内容。

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

* 简单变更（**小修小改**）：直接用“/opsx:propose”，省时间，一键到位；
* 复杂变更（**大功能 / 重构**）：先用“/opsx:new”搭架子，再用“/opsx:continue”分步完善，流程更可控；
* 核心差异：propose 是「**提想法 + 做规划**」，new 只是「**建空架子**」。

## 使用openspec

**实战项目：**[https://github.com/eyjian/github-buddy](https://github.com/eyjian/github-buddy)

### 构建文档

执行“/opsx:propose”开始一个新需求。提示词：**根据文件ai_requirements.md的内容执行，放在/data/workspace/github.com/eyjian/github-buddy目录下**。完成后可到openspec/changes/github-buddy目录下查看生成的文件。

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

在这个过程中，如何AI设计的不符合预期，可对话方式优化，比如开发语言采用了C++17，需要改为Go1.21+。所有问题都确认好后，即可执行“/opsx:apply”驱动开发。

### 编写代码

执行“**/opsx:apply**”，AI即开始根据specs目录下的文件，生成对应的代码。AI会安装**task.md**中的任务，逐步完成代码实现。如下是部分任务实现过程：

```
好的！状态是 ready，41 个任务，0 个已完成。现在读取所有上下文文件来开始实现。proposal.md 已在附件中，让我读取其余的上下文文件：

读取文件
design.md
读取文件
tasks.md
读取文件
spec.md
读取文件
spec.md
读取文件
spec.md
读取文件
spec.md
读取文件
spec.md
所有上下文文件已读取完毕。让我先查看当前项目目录结构：

查找目录
github-buddy
好的，项目目前只有 doc/ 和 openspec/ 目录，没有任何代码文件。从零开始！

Implementing: github-buddy (schema: spec-driven)
Progress: 0/41 tasks complete
Remaining: 41 tasks

现在开始逐步实现。先从任务 1.1 开始：

Working on task 1/41: 1.1 初始化 Go Module 项目结构
执行命令
cd /data/workspace/github.com/eyjian/github-buddy && go mod init github.com/eyjian/github-buddy && mkdir -p cmd/github-buddy internal/platform internal/storage internal/detector internal/hosts internal/backup internal/cache internal/config internal/logger pkg
现在创建 main.go 入口文件：

编辑文件
main.go
创建根命令占位文件（将在任务 1.2 中完善）：

编辑文件
root.go
现在标记任务 1.1 完成，同时开始安装依赖（任务 1.2）：

执行命令
cd /data/workspace/github.com/eyjian/github-buddy && go get github.com/spf13/cobra@latest github.com/rs/zerolog@latest gopkg.in/natefinch/lumberjack.v2@latest && go mod tidy
依赖安装成功。验证编译是否通过：

执行命令
cd /data/workspace/github.com/eyjian/github-buddy && go build ./cmd/github-buddy/ && echo "BUILD SUCCESS"
编译成功！现在标记任务 1.1 和 1.2 完成，然后继续实现任务 1.3（平台抽象层）和 1.4（存储模块）：

编辑文件
tasks.md
✓ Task 1.1 complete
✓ Task 1.2 complete

。。。。。。

Working on tasks 25-30/41: CLI 命令实现
现在实现日志系统（任务 7.1-7.3，CLI 命令依赖日志），然后是 CLI 命令。

编辑文件
logger.go
现在实现 CLI 命令。首先更新 root.go 加入日志初始化和全局参数：

编辑文件
root.go

现在实现 init 命令：

编辑文件
cmd_init.go
编辑文件
cmd_update.go
编辑文件中
```

在这个“**/opsx:apply**”过程中，AI不但自动编写代码，还会自动编译代码，同时还好对程序进行测试，确保代码质量，这是一个迭代过程，但都是AI在吭哧吭哧的干，基本不需要人工参与。

在“/opsx:apply”完成后，会给出执行总结：

**Implementation Complete**

**Change**: github-buddy
**Schema**: spec-driven
**Progress**: 41/41 tasks complete ✓

**已完成的 9 大模块**

| 阶段       | 模块                                  | 文件                                                                 |
|------------|---------------------------------------|----------------------------------------------------------------------|
| 1. 项目基础 | Go Module + cobra + zerolog           | cmd/github-buddy/main.go, root.go, go.mod                            |
| 2. IP 检测  | 数据源获取 + ICMP/TCP 检测 + 评分筛选 | internal/detector/ (6 个文件)                                        |
| 3. Hosts 管理 | 读取/写入/标记区块/冲突检测           | internal/hosts/ (4 个文件)                                           |
| 4. 备份回滚 | 自动备份 + SHA256 校验 + 回滚恢复    | internal/backup/backup.go                                            |
| 5. 缓存配置 | IP 缓存 + 过期检查 + 配置管理         | internal/cache/ + internal/config/                                   |
| 6. CLI 命令 | init update rollback / status         | cmd/github-buddy/cmd_*.go (4 个文件)                                 |
| 7. 日志系统 | zerolog + lumberjack 双后端           | internal/logger/logger.go                                            |
| 8. 测试     | 4 个 _test.go 文件，全部通过          | detector hosts backup / platform                                     |
| 9. 构建分发 | Makefile + GitHub Actions + Goreleaser | Makefile, .github/, .goreleaser.yml                                  |

**四个核心命令**

* **github-buddy init** — 初始化目录、首次检测最优 IP、更新 hosts

* **github-buddy update** — 强制重新检测、备份并更新 hosts

* **github-buddy rollback** — 从备份恢复 hosts 文件

* **github-buddy status** — 实时检测并展示当前状态表格

所有任务已完成！你可以使用 **/opsx:archive** 来归档这个变更。

可以看到AI输出的结构很标准漂亮：

```
[.../workspace/github.com/eyjian/github-buddy]# tree .
.
├── cmd
│   └── github-buddy
│       ├── cmd_init.go
│       ├── cmd_rollback.go
│       ├── cmd_status.go
│       ├── cmd_update.go
│       ├── main.go
│       └── root.go
├── doc
│   ├── ai_requirements.md
│   ├── Getting_Started_with_OpenSpec.md
│   └── manual_requirements.md
├── github-buddy
├── go.mod
├── go.sum
├── internal
│   ├── backup
│   │   ├── backup.go
│   │   └── backup_test.go
│   ├── cache
│   │   └── cache.go
│   ├── config
│   │   └── config.go
│   ├── detector
│   │   ├── defaults.go
│   │   ├── detector.go
│   │   ├── detector_test.go
│   │   ├── ping.go
│   │   ├── scorer.go
│   │   ├── source.go
│   │   ├── tcp.go
│   │   └── types.go
│   ├── hosts
│   │   ├── block.go
│   │   ├── conflict.go
│   │   ├── hosts_test.go
│   │   ├── reader.go
│   │   └── writer.go
│   ├── logger
│   │   └── logger.go
│   ├── platform
│   │   ├── platform.go
│   │   └── platform_test.go
│   └── storage
│       └── storage.go
├── Makefile
├── openspec
│   └── changes
│       └── github-buddy
│           ├── design.md
│           ├── proposal.md
│           ├── specs
│           │   ├── auto-update
│           │   │   └── spec.md
│           │   ├── backup-rollback
│           │   │   └── spec.md
│           │   ├── cli-commands
│           │   │   └── spec.md
│           │   ├── hosts-management
│           │   │   └── spec.md
│           │   └── ip-detection
│           │       └── spec.md
│           └── tasks.md
└── pkg

22 directories, 42 files
```

AI自动生成的命令标准专业，中英文混合了，这个可通过对话调整为英文，也可通过设置rule方式约束好，省去每次都要调整的麻烦。

```
[.../workspace/github.com/eyjian/github-buddy]# ./github-buddy -h
GitHub-Buddy 通过智能维护 GitHub 域名-IP 映射并修改系统 hosts 文件，
解决开发者执行 go mod tidy 和 git clone 时的网络访问问题。

Usage:
  github-buddy [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        初始化 github-buddy
  rollback    回滚 hosts 文件
  status      查看当前状态
  update      强制更新 GitHub IP

Flags:
  -f, --force     跳过确认提示，强制执行
  -h, --help      help for github-buddy
  -v, --verbose   输出详细日志到控制台
      --version   version for github-buddy

Use "github-buddy [command] --help" for more information about a command.
```
