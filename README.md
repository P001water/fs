# fs


---

你可以将此分支看作fscan的活跃实验分支，区别就是结合个人需求和代码习惯，开发命名习惯不同修改太多，功能成熟到时候精简代码向fscan提pr。

当前修改包括

* 优化存活主机输出顺序 - 按顺序输出

* 优化端口探测结果输出 - 如下图
* 修改输出文件名 - r.txt
* 部分功能删除 - 例如JSON格式输出等
* 其他逻辑修改等等

注：fscan 1.8.4的releases在win7等平台运行报错

go 从 1.21版本放弃了对Windows 全平台的支持，，编译go版本建议使用go 1.20

参考：[Go 1.20 Release Notes - The Go Programming Language](https://go.dev/doc/go1.20#windows)

tips: 360不杀 火绒fscan特征消除（代码里藏有免杀的密码）



##  使用演示

其他大致使用方法和fscan差不多

```
fs -h 101.43.3.85-100
```

![image-20240604192710413](./img/image-20240604192710413.png)
