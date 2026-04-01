执行“go mod tidy”和“git clone”时，都存在访问github.com的网络问题，需要开发一个工具，用于维护github.com的域名同IP地址间的映射，以解决执行“go mod tidy”和“git clone”时遭遇的网络问题。工具采用修改/etc/hosts的方式来达到目的，在保证网络可通的前提下，还要保证访问速度流畅不卡顿。

以上内容用于构建ai_requirements.md，由豆包生成ai_requirements.md文件。
