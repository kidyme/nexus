# IDEA

## 基本引擎

https://github.com/gorse-io/gorse
gorse: 一个基于人工智能的开源推荐系统，使用 Go 语言编写。Gorse 的目标是成为一个通用的开源推荐系统，能够快速集成到各种在线服务中。通过将商品、用户和交互数据导入 Gorse，系统会自动训练模型，为每个用户生成推荐
go实现，start9.3k

社区博客:
https://cloud.tencent.com/developer/article/2064808

https://github.com/metarank/metarank
metarank: 是一款开源排名服务。它可以帮助您构建个性化的语义/神经搜索和推荐系统
scala实现，start2.4k

## 应用

https://github.com/gorse-io/gitrec
GitRec 是基于 Gorse 的 GitHub 仓库推荐系统，是目前缺失的推荐系统。
> 可以考虑做一个线上实际的推荐系统（小红书-博主侧的素材推荐）

## 生产级框架

https://help.aliyun.com/zh/airec/what-is-pai-rec/product-overview/what-is-pairec?spm=a2c4g.11186623.0.i0
https://github.com/alibaba/pairec (go接入pairec的demo)

阿里的推荐系统开发平台PAI-Rec
![PAI-Rec](./assets/PAI-Rec.png)

无源码，无法根据其重构，但自己尝试实现其具有的功能:
模型：数据标注、模型构建、模型训练、模型部署、推理优化
全链路的需求: 召回、粗排、排序、重排、多目标和冷启动
特征处理、样本生成、画像管理、模型调度、数据更新
数仓引擎: 海量(可能海不了，这块不行就借助云服务或者开源中间件)数据实时写入、实时更新、实时加工和实时分析