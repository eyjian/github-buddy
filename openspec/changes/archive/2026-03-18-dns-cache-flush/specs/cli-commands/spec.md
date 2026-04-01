## MODIFIED Requirements

### REQ-1: init 命令
- 命令：`github-buddy init`
- 功能：
  1. 创建 `~/.github-buddy/` 目录结构（config/cache/logs/backups）
  2. 生成默认配置文件 `config.json`
  3. 备份当前 hosts 文件
  4. 执行首次 IP 检测
  5. 将最优 IP 写入 hosts 文件
  6. **自动刷新系统 DNS 缓存并输出提示**
  7. 输出初始化结果摘要
- 已初始化时重复执行，提示"已初始化，可使用 update 更新 IP"

#### Scenario: init 成功后刷新 DNS 缓存
- **WHEN** init 命令成功将最优 IP 写入 hosts 文件
- **THEN** 系统 SHALL 自动调用 DNS 缓存刷新，输出刷新结果提示，然后输出初始化结果摘要

#### Scenario: init 成功但 DNS 刷新失败
- **WHEN** init 命令成功更新 hosts 但 DNS 缓存刷新失败
- **THEN** 系统 SHALL 输出 DNS 刷新警告但 init 命令整体仍返回成功

### REQ-2: update 命令
- 命令：`github-buddy update`
- 功能：
  1. 强制执行完整 IP 检测（忽略缓存）
  2. 备份当前 hosts 文件
  3. 将最优 IP 写入 hosts 文件
  4. 更新缓存
  5. **自动刷新系统 DNS 缓存并输出提示**
  6. 输出更新结果（域名-新IP-旧IP-延迟 对比表）

#### Scenario: update 成功后刷新 DNS 缓存
- **WHEN** update 命令成功将最优 IP 写入 hosts 文件
- **THEN** 系统 SHALL 自动调用 DNS 缓存刷新，输出刷新结果提示，然后输出更新结果摘要

#### Scenario: update 成功但 DNS 刷新失败
- **WHEN** update 命令成功更新 hosts 但 DNS 缓存刷新失败
- **THEN** 系统 SHALL 输出 DNS 刷新警告但 update 命令整体仍返回成功
