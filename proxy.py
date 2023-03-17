import winreg

class ProxyServer:
 
    def __init__(self):
        self.__path = r'Software\Microsoft\Windows\CurrentVersion\Internet Settings'
        self.__INTERNET_SETTINGS = winreg.OpenKeyEx(winreg.HKEY_CURRENT_USER, self.__path, 0, winreg.KEY_ALL_ACCESS)
 
    def get_server_form_Win(self):
        ip, port = "", ""
        if self.is_open_proxy_form_Win():
            try:
                ip, port = winreg.QueryValueEx(self.__INTERNET_SETTINGS, "ProxyServer")[0].split(":")
                print("获取到代理信息：{}:{}".format(ip, port))
            except FileNotFoundError as err:
                print("没有找到代理信息：" + str(err))
            except Exception as err:
                print("有其他报错：" + str(err))
        else:
            print("系统没有开启代理")
        return ip, port
 
    def is_open_proxy_form_Win(self):
        """判断是否开启了代理"""
        try:
            if winreg.QueryValueEx(self.__INTERNET_SETTINGS, "ProxyEnable")[0] == 1:
                return True
        except FileNotFoundError as err:
            print("没有找到代理信息：" + str(err))
        except Exception as err:
            print("有其他报错：" + str(err))
        return False