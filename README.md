WSocks-GO
=======
这里是WSocks的go版本。  
之后应该不会再更新WSocks的Kotlin版本了，这种小工具跑在JVM上还是太脑瘫了一点。  

Server
----
Release文件
> Binary： <a href="https://github.com/Wooyme/WSocks-Go/releases/download/v1.0/wsocks-server-linux-amd64">Release</a>  

运行
> ./wsocks-server config.json

配置
> {"port": 1889,
   "users": [
     {
       "user": "username",
       "pass": "password",
       "multiple": -1,
       "limit": -1
     }
   ]
 }
 
删掉了offset，省略了zip。Go的版本就不考虑兼容之前Kotlin版本了。

Client
----
Release文件
> Binary <a href="https://github.com/Wooyme/WSocks-Go/releases/tag/v1.0">Release</a>

Windows版本
> 自带PAC模式和全局模式,可以在托盘里选择,实测IE系列有点问题。Firefox和Chrome都正常

Linux版本
> Linux发行版太多了，自己给浏览器装插件吧。

通用
> 按照国际惯例,本地端口开在1080。要是不巧1080被占用了....自己改一下重新编译.
> 程序启动后会打开默认浏览器。为了省事，就不写GUI了。Go的GUI挺搞人心态的，而且GUI也不方便跨平台。

vs WSocks Kolin .ver
----------------
迫使我从Kotlin迁移到Go的原因其实主要是因为JVM上有内存泄露，但是我又找不到在哪里。
初步推断是Vert.x的websocket有点问题，但是看了半天也没找到问题在哪。于是干脆换个平台换个心情。  

相较于WSocks-Kt，WSocks-Go占用更少的内存。WSocks-Kt启动后至少占用100M左右的内存，再加上客户端泄露，
运行一段时间后，可能达到300M的内存。WSocks-Go启动内存在20M左右，并且在运行期间可以一直保持这个内存占用量。

而且Go编译后的文件体积也更小。WSocks-Kt为了搞个Windows的安装包，得把整个JRE打进去，压缩后体积也到了60M。

