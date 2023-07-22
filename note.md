我计划实现这个功能. 先fork一个分支尝试一下.

我的大概计划是
* 数据库里新建 `tags` 表(tag元信息), `entry_tags`表(关联entry和tag). entries表里已经有了`tags`字段, 这个数据是从 RSS 的 category 获取的, 这里可能会导致歧义.

* 在 unread, history, feeds/entry 这几个显示文章的节目加上添加 tag 按钮, 弹出模态框

* 新增 tags 页面, 管理方式类似现有是 categories, 新建tag, 查询tag

* tag 支持分层, 实现上打算简单讲层级用 `/` 分割. `tech/rust` 类似这样为tag命名在tags界面会分层显示.


## TODO

原本程序了已经有了 Tag 的概念([]text), 包括 Tag 过滤, 这个重名很难搞

## TODO

1. 新建移除更新ctag添加校验, 参考categories的校验.
    ctags有嵌套结构, 这个也需要校验.

2. /ctag/{:ctagID}/entry/{:entryID} 界面, 需要处理分页


## TODO

* rclone 的密钥盐必须被记住, 不能使用随机生成的

