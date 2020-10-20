##一个根据2018年之前豆瓣电影数据库来格式化本地电影文件名的小工具

###使用方法

#####./movie_name_tidy.exe run \<option>

###输入参数

#####-dir 指定目录 -dir=example
#####-name_format 指定文件名格式  -name_format=[year][star]title
    一般字符串会被保留，关键字会被替换
    year  年份
    star  评分
    director 导演
    title_short 中文名
    title 中外全名
#####-manual_mode 确认模式 -manual_mode=false
    默认true 每次重命名都要确认
    如果false 没有重名的情况下会自动重命名
#####-recent_file_count 只重命名指定数量的最近文件
    #todo
#####-filename 指定原文件名
    -filename=XXX