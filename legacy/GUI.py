import logging

from PyQt5.QtGui import QIcon
from PyQt5.QtWidgets import QApplication, QWidget, QVBoxLayout, QLabel, QLineEdit, QPushButton, QMessageBox, QListWidget, QTextEdit
from PyQt5.QtWebEngineWidgets import QWebEngineView
from PyQt5.QtCore import QUrl, Qt
import resources

from PyQt5 import QtCore, QtGui, QtWidgets

import stat
import sys
import ctypes
import zlib
import base64
import json
import random
import string
import psutil
import subprocess
import shutil
import winreg
import os
import time
import requests
import webbrowser
import win32crypt
import vdf
from jwt_helper import verify_steam_jwt

CURRENT_VERSION = "2.0.4"
VERSION_CHECK_URL = "https://login-software-api.sdaewrcawre.com/version.json"


def get_steam_install_path():
    steam_pid = get_pid('steam.exe')
    if steam_pid != 0:
        process = psutil.Process(steam_pid)
        steam_path = process.exe()
        subprocess.run('taskkill /f /im steam.exe', shell=True)
        subprocess.run('taskkill /f /im steamwebhelper.exe', shell=True)
        time.sleep(2)
    else:
        steam_path = read_registry_value('Software\\Classes\\steam\\Shell\\Open\\Command', '').replace('"', '')
        steam_path = steam_path[:len(steam_path) - 6]
    return steam_path[:len(steam_path) - 9]


def get_local_vdf_path():
    app_data_path = os.getenv('localappdata')
    if app_data_path is None:
        logger.error("Failed to get localappdata environment variable.")
    return os.path.join(app_data_path, 'Steam', 'local.vdf')


def get_all_login_users():
    path = get_steam_install_path()
    if not os.path.exists(path):
        QMessageBox.critical(None, "Error", "Directory not recognized, please open the client in advance!")
        return 0
    path = os.path.join(path, 'config', 'loginusers.vdf')
    if not os.path.exists(path):
        return []
    result = []
    with open(path, 'r', encoding='utf-8') as f:
        data = vdf.load(f)
        for key in data.get('users'):
            result.append(data.get('users').get(key).get('AccountName'))
    return result


def steam_decrypt(encrypted_hex_data):
    encrypted_data = bytes.fromhex(encrypted_hex_data)
    decrypted_data = win32crypt.CryptUnprotectData(encrypted_data, None, None, None, 0)
    return decrypted_data[1].decode('utf-8')


def add_to_current_cache(account_name, eya):
    save_current_cache()
    current_users = {}
    if os.path.exists('user_backup.json'):
        with open('user_backup.json', 'r') as f:
            current_users = json.load(f)
    current_users[account_name] = eya
    with open('user_backup.json', 'w') as f:
        json.dump(current_users, f)


def remove_readonly(func, path, excinfo):
    if os.path.exists(path):
        try:
            os.chmod(path, stat.S_IWRITE)
        except Exception:
            pass
        func(path)


def save_current_cache():
    local_path = get_local_vdf_path()
    all_login_users = get_all_login_users()
    if not all_login_users:
        return
    account_name_mapping = {compute_crc32(account_name) + "1": account_name for account_name in all_login_users}
    current_login_users = {}
    if os.path.exists(local_path):
        try:
            with open(local_path, 'r', encoding='utf-8') as f:
                data = vdf.load(f)
                cache = data.get('MachineUserConfigStore').get('Software').get('Valve').get('Steam').get('ConnectCache')
                for key in cache:
                    try:
                        jwt_encrypted = cache[key]
                        account_name = account_name_mapping.get(key)
                        decrypted = steam_decrypt(jwt_encrypted)
                        current_login_users[account_name] = decrypted
                    except Exception:
                        pass
        except Exception:
            pass
    current_users = {}
    if os.path.exists('user_backup.json'):
        try:
            with open('user_backup.json', 'r') as f:
                current_users = json.load(f)
        except Exception:
            pass
    current_users.update(current_login_users)
    with open('user_backup.json', 'w') as f:
        json.dump(current_users, f)


def steam_encrypt(token, account_name):
    data_to_encrypt = token.encode('utf-8')
    byte_string = b'B\x00O\x00b\x00f\x00u\x00s\x00c\x00a\x00t\x00e\x00B\x00u\x00f\x00f\x00e\x00r\x00\x00\x00'
    encrypted_data = win32crypt.CryptProtectData(data_to_encrypt, byte_string.decode('utf-8'), account_name.encode('utf-8'), None, None, 17)
    return encrypted_data.hex()


def source_path(relative_path):
    if getattr(sys, 'frozen', False):
        base_path = sys._MEIPASS
    else:
        base_path = os.path.abspath(".")
    return os.path.join(base_path, relative_path)


def read_registry_value(key_path, value_name):
    key = winreg.OpenKey(winreg.HKEY_CURRENT_USER, key_path, 0, winreg.KEY_READ)
    value, _ = winreg.QueryValueEx(key, value_name)
    winreg.CloseKey(key)
    if isinstance(value, bytes):
        value = value.decode('utf-8')
    return value


def get_pid(process_name):
    for process in psutil.process_iter():
        if process.name() == process_name:
            return process.pid
    return 0


def compute_crc32(data):
    crc32_value = zlib.crc32(data.encode('utf-8'))
    return f'{crc32_value:08x}'.lstrip('0')


def parse_eya(eya):
    token_arr = eya.split('.')
    if len(token_arr) != 3:
        QMessageBox.critical(None, "Error", "Invalid token!")
        return None
    padding = len(token_arr[1]) % 4
    if padding != 0:
        token_arr[1] += '=' * (4 - padding)
    return base64.b64decode(token_arr[1]).decode('utf-8')


def build_config(mtbf, steam_id, account_name):
    config = {
        "InstallConfigStore": {
            "Software": {
                "Valve": {
                    "Steam": {
                        "AutoUpdateWindowEnabled": "0",
                        "ipv6check_http_state": "bad",
                        "ipv6check_udp_state": "bad",
                        "ShaderCacheManager": {
                            "HasCurrentBucket": "1",
                            "CurrentBucketGPU": "b4799b250d4196b0;36174e7cc31a08f9",
                            "CurrentBucketDriver": "W2:c18b09d9c69329b41cdbbf3de627bc9f;W2:ee32edf67d134b7cc2ec0cdecbd00037"
                        },
                        "RecentWebSocket443Failures": "",
                        "RecentWebSocketNon443Failures": "",
                        "RecentUDPFailures": "",
                        "RecentTCPFailures": "",
                        "Accounts": {
                            account_name: {
                                "SteamID": steam_id
                            }
                        },
                        "CellIDServerOverride": "170",
                        "MTBF": mtbf,
                        "cip": "02000000507a6c24d6e96c6b00004021a356",
                        "SurveyDate": "2017-10-22",
                        "SurveyDateVersion": "-1724767764117155760",
                        "SurveyDateType": "3",
                        "Rate": "30000"
                    }
                }
            },
            "SDL_GamepadBind": {
                "03000000de280000ff11000001000000,Steam Virtual Gamepad": "a:b0,b:b1,back:b6,dpdown:h0.4,dpleft:h0.8,dpright:h0.2,dpup:h0.1,leftshoulder:b4,leftstick:b8,lefttrigger:+a2,leftx:a0,lefty:a1,rightshoulder:b5,rightstick:b9,righttrigger:-a2,rightx:a3,righty:a4,start:b7,x:b2,y:b3,platform:Windows",
                "03000000de280000ff11000000000000,Steam Virtual Gamepad": "a:b0,b:b1,back:b6,dpdown:h0.4,dpleft:h0.8,dpright:h0.2,dpup:h0.1,leftshoulder:b4,leftstick:b8,lefttrigger:+a2,leftx:a0,lefty:a1,rightshoulder:b5,rightstick:b9,righttrigger:-a2,rightx:a3,righty:a4,start:b7,x:b2,y:b3",
                "03000000de280000ff11000000007701,Steam Virtual Gamepad": "a:b0,b:b1,x:b2,y:b3,back:b6,start:b7,leftstick:b8,rightstick:b9,leftshoulder:b4,rightshoulder:b5,dpup:b10,dpdown:b12,dpleft:b13,dpright:b11,leftx:a1,lefty:a0~,rightx:a3,righty:a2~,lefttrigger:a4,righttrigger:a5,"
            },
            "streaming": {
                "ClientID": "-6167702798309564492"
            },
            "Music": {
                "LocalLibrary": {
                    "Directories": {
                        "0": "0200000013500d50b7e96d1b621bcb56f5a12ce5e0651b4c3b3a50a59063d16c7d6fc334d903be2347030590f9d63c5a09370ac77bcfc6c945d0b348b91a586438e4162d56e494c9c73173ae",
                        "1": "0200000013500d50b7e96d1b621bcb56f2a12ce3e76508523b7812f4c237ec51786ff63aac72a1212b7416a2f0d71b7039302dcc2ebca3bb45d2c02bd87645623d8a784832e49cc1a01779a2209be04c"
                    }
                }
            }
        }
    }
    return vdf.dumps(config)


def build_login_users(steam_id, account_name):
    login_users = {
        "users": {
            steam_id: {
                "AccountName": account_name,
                "PersonaName": account_name,
                "RememberPassword": "1",
                "WantsOfflineMode": "0",
                "SkipOfflineModeWarning": "0",
                "AllowAutoLogin": "1",
                "MostRecent": "1",
                "Timestamp": str(int(time.time()))
            }
        }
    }
    return vdf.dumps(login_users)


def build_local(crc32, jwt):
    local = {
        "MachineUserConfigStore": {
            "Software": {
                "Valve": {
                    "Steam": {
                        "ConnectCache": {
                            crc32: jwt
                        }
                    }
                }
            }
        }
    }
    return vdf.dumps(local)


def reset_steam():
    path = get_steam_install_path()
    for directory in [os.path.join(path, 'userdata'), os.path.join(path, 'config')]:
        if os.path.exists(directory):
            shutil.rmtree(directory, onerror=remove_readonly)
    local_vdf_path = get_local_vdf_path()
    if os.path.exists(local_vdf_path):
        os.remove(local_vdf_path)
    subprocess.Popen(os.path.join(path, 'steam.exe'), shell=True)
    QMessageBox.information(None, "Reset Completed", "Steam has been reset successfully.")


def login_game(eya, account_name):
    if '@' in account_name:
        # Remove everything after @
        account_name = account_name.split('@')[0]
    save_current_cache()
    crc32_account_name = compute_crc32(account_name) + "1"
    json_data = json.loads(parse_eya(eya))
    if not json_data:
        return
    steam_id = json_data['sub']
    mtbf = ''.join(random.choices(string.digits, k=9))
    jwt = steam_encrypt(eya, account_name)
    path = get_steam_install_path()
    local_vdf_path = get_local_vdf_path()
    if os.path.exists(local_vdf_path):
        os.remove(local_vdf_path)
    try:
        os.makedirs(os.path.join(path, 'config'))
    except FileExistsError:
        pass
    key = winreg.OpenKey(winreg.HKEY_CURRENT_USER, 'SOFTWARE\\Valve\\Steam', 0, winreg.KEY_WRITE)
    winreg.SetValueEx(key, 'AutoLoginUser', 0, winreg.REG_SZ, account_name)
    winreg.CloseKey(key)
    config = build_config(mtbf, steam_id, account_name)
    login_users = build_login_users(steam_id, account_name)
    local = build_local(crc32_account_name, jwt)
    remove_readonly(os.remove, os.path.join(path, 'config', 'config.vdf'), None)
    with open(os.path.join(path, 'config', 'config.vdf'), 'w', encoding='utf-8') as f:
        f.write(config)
    remove_readonly(os.remove, os.path.join(path, 'config', 'loginusers.vdf'), None)
    with open(os.path.join(path, 'config', 'loginusers.vdf'), 'w', encoding='utf-8') as f:
        f.write(login_users)
    local_vdf_path = get_local_vdf_path()
    if os.path.exists(local_vdf_path):
        remove_readonly(os.remove, local_vdf_path, None)
        os.remove(local_vdf_path)
    with open(local_vdf_path, 'w', encoding='utf-8') as f:
        f.write(local)
    add_to_current_cache(account_name, eya)
    subprocess.Popen(os.path.join(path, 'steam.exe'), shell=True)
    QMessageBox.information(None, "Success", "Login initiated successfully.")


def compare_versions(current_version, latest_version):
    current_version_parts = list(map(int, current_version.split('.')))
    latest_version_parts = list(map(int, latest_version.split('.')))

    for current_part, latest_part in zip(current_version_parts, latest_version_parts):
        if current_part < latest_part:
            return -1
        elif current_part > latest_part:
            return 1

    return 0


def check_for_updates():
    try:
        response = requests.get(VERSION_CHECK_URL)
        response.raise_for_status()
        version_info = response.json()
        latest_version = version_info["latest_version"]
        change_log = version_info["change_log"]
        download_url = version_info["download_url"]
        force_update_below_version = version_info["force_update_below_version"]

        comparison_result = compare_versions(CURRENT_VERSION, latest_version)
        force_update_comparison = compare_versions(CURRENT_VERSION, force_update_below_version)

        if force_update_comparison <= 0:
            msg_box = QMessageBox()
            msg_box.setIcon(QMessageBox.Information)
            msg_box.setWindowTitle("Update Required")
            msg_box.setText(f"A mandatory update ({latest_version}) is available.\n\nChange Log:\n{change_log}")
            msg_box.setStyleSheet("background-color: #2b2b2b; color: #FFFFFF;")
            msg_box.exec_()
            webbrowser.open(download_url)
            sys.exit()
        elif comparison_result < 0:
            msg_box = QMessageBox()
            msg_box.setIcon(QMessageBox.Question)
            msg_box.setWindowTitle("Update Available")
            msg_box.setText(f"A new version ({latest_version}) is available.\n\nChange Log:\n{change_log}\n\nDo you want to update now?")
            msg_box.setStandardButtons(QMessageBox.Yes | QMessageBox.No)
            msg_box.setStyleSheet("background-color: #2b2b2b; color: #FFFFFF;")
            if msg_box.exec_() == QMessageBox.Yes:
                webbrowser.open(download_url)
        else:
            msg_box = QMessageBox()
            msg_box.setIcon(QMessageBox.Information)
            msg_box.setWindowTitle("Update Check")
            msg_box.setText("You are using the latest version.")
            msg_box.setStyleSheet("background-color: #2b2b2b; color: #FFFFFF;")
            msg_box.exec_()
    except Exception as e:
        logger.error(f"Failed to check for updates: {e}", exc_info=True)
        msg_box = QMessageBox()
        msg_box.setIcon(QMessageBox.Warning)
        msg_box.setWindowTitle("Update Check Failed")
        msg_box.setText("Failed to check for updates.")
        msg_box.setStyleSheet("background-color: #2b2b2b; color: #FFFFFF;")
        msg_box.exec_()



class Ui_MainWindow(QtWidgets.QMainWindow):
    def __init__(self):
        super().__init__()

        self.setup_ui(self)

    def load_accounts(self):
        self.AccountsList.clear()
        if os.path.exists('user_backup.json'):
            try:
                with open('user_backup.json', 'r') as f:
                    self.current_users = json.load(f)
                    for account in self.current_users:
                        self.AccountsList.addItem(account)
            except Exception as e:
                QMessageBox.warning(None, "Error", f"Failed to load accounts: {e}")

    def setup_ui(self, MainWindow):
        MainWindow.setObjectName("MainWindow")
        MainWindow.setFixedSize(1000, 700)
        MainWindow.setStyleSheet("background-color: rgb(0, 0, 0);\n"
                                 "color: rgb(255, 255, 255);")
        MainWindow.setIconSize(QtCore.QSize(24, 24))
        MainWindow.setDocumentMode(False)
        self.centralwidget = QtWidgets.QWidget(MainWindow)
        self.centralwidget.setObjectName("centralwidget")
        self.Background = QtWidgets.QFrame(self.centralwidget)
        self.Background.setGeometry(QtCore.QRect(0, 0, 1000, 700))
        self.Background.setStyleSheet("background-color: #171717;")
        self.Background.setFrameShape(QtWidgets.QFrame.StyledPanel)
        self.Background.setFrameShadow(QtWidgets.QFrame.Raised)
        self.Background.setObjectName("Background")
        self.AccountKey = QtWidgets.QLineEdit(self.Background)
        self.AccountKey.setGeometry(QtCore.QRect(79, 136, 360, 79))
        font = QtGui.QFont()
        font.setFamily("Microsoft YaHei")
        font.setPointSize(10)
        font.setBold(True)
        font.setWeight(75)
        self.AccountKey.setFont(font)
        self.AccountKey.setStyleSheet("background-image: url(:/Textbox/Users/Administrator/Downloads/lineedit.png);\n"
                                      "background: #1d1d1d;\n"
                                      "border-radius: 24px;\n"
                                      "color: #bab1b1;\n"
                                      "border: 3px solid #1d1d1d;\n"
                                      "outline: none;\n"
                                      "padding-left: 10px;\n"
                                      "padding: 25%;")
        self.AccountKey.setObjectName("AccountKey")
        self.AccountKey.setFrame(False)
        self.AccountKey.setAcceptDrops(False)
        self.AccountKey.setAttribute(Qt.WA_MacShowFocusRect, 0)
        self.LoginButton = QtWidgets.QPushButton(self.Background)
        self.LoginButton.setGeometry(QtCore.QRect(77, 242, 363, 62))
        self.LoginButton.clicked.connect(self.on_login)
        self.LoginButton.setFont(font)
        self.LoginButton.setCursor(QtGui.QCursor(QtCore.Qt.PointingHandCursor))
        self.LoginButton.setStyleSheet("background: linear-gradient(90deg, rgba(240,0,0,1) 0%, rgba(220,40,30,1) 100%);\n"
                                       "background-color: qlineargradient(spread:pad, x1:0, y1:0, x2:1, y2:0, stop:0 rgba(240, 0, 0, 255), stop:1 rgba(220, 40, 30, 255));\n"
                                       "border-radius: 24px;\n"
                                       "color: rgb(20,20,20)")
        self.LoginButton.setAutoDefault(False)
        self.LoginButton.setFlat(False)
        self.LoginButton.setObjectName("LoginButton")

        # Change the geometry settings to ensure AccountsList is within bounds
        self.AccountsList = QtWidgets.QListWidget(self.Background)
        self.AccountsList.setGeometry(QtCore.QRect(78, 440, 362, 151))
        self.AccountsList.setStyleSheet("background-image: url(:/AccountsList/Users/Administrator/Downloads/accountslist.png);\n"
                                        "border: 0;\n"
                                        "padding: 15%;\n")
        self.AccountsList.setObjectName("AccountsList")
        self.AccountsList.itemClicked.connect(self.on_account_selected)


        self.RestoreAccountButton = QtWidgets.QPushButton(self.Background)
        self.RestoreAccountButton.clicked.connect(self.on_restore_account)
        self.RestoreAccountButton.setGeometry(QtCore.QRect(78, 601, 172, 44))
        self.RestoreAccountButton.setFont(font)
        self.RestoreAccountButton.setCursor(QtGui.QCursor(QtCore.Qt.PointingHandCursor))
        self.RestoreAccountButton.setStyleSheet("background: linear-gradient(90deg, rgba(240,0,0,1) 0%, rgba(220,40,30,1) 100%);\n"
                                        "background-color: qlineargradient(spread:pad, x1:0, y1:0, x2:1, y2:0, stop:0 rgba(240, 0, 0, 255), stop:1 rgba(220, 40, 30, 255));\n"
                                        "border-radius: 22px;\n"
                                        "color: rgb(20,20,20)")
        self.RestoreAccountButton.setObjectName("RestoreAccountButton")
        self.DeleteButton = QtWidgets.QPushButton(self.Background)
        self.DeleteButton.setGeometry(QtCore.QRect(267, 601, 172, 44))
        self.DeleteButton.setFont(font)
        self.DeleteButton.setCursor(QtGui.QCursor(QtCore.Qt.PointingHandCursor))
        self.DeleteButton.setStyleSheet("background: linear-gradient(90deg, rgba(240,0,0,1) 0%, rgba(220,40,30,1) 100%);\n"
                                        "background-color: qlineargradient(spread:pad, x1:0, y1:0, x2:1, y2:0, stop:0 rgba(240, 0, 0, 255), stop:1 rgba(220, 40, 30, 255));\n"
                                        "border-radius: 22px;\n"
                                        "color: rgb(20,20,20)")
        self.DeleteButton.setObjectName("DeleteButton")
        self.DeleteButton.clicked.connect(self.on_delete_account)
        self.ResetSteamButton = QtWidgets.QPushButton(self.Background)
        self.ResetSteamButton.setGeometry(QtCore.QRect(559, 595, 363, 50))
        self.ResetSteamButton.setFont(font)
        self.ResetSteamButton.clicked.connect(reset_steam)
        self.ResetSteamButton.setCursor(QtGui.QCursor(QtCore.Qt.PointingHandCursor))
        self.ResetSteamButton.setStyleSheet("background: linear-gradient(90deg, rgba(240,0,0,1) 0%, rgba(220,40,30,1) 100%);\n"
                                          "background-color: qlineargradient(spread:pad, x1:0, y1:0, x2:1, y2:0, stop:0 rgba(240, 0, 0, 255), stop:1 rgba(220, 40, 30, 255));\n"
                                          "border-radius: 22px;\n"
                                          "color: rgb(20,20,20)")
        self.ResetSteamButton.setObjectName("ResetSteamButton")

        self.textBrowser = QWebEngineView(self.centralwidget)
        self.textBrowser.setWindowFlags(QtCore.Qt.FramelessWindowHint)
        self.textBrowser.setGeometry(QtCore.QRect(543, 111, 392, 387))
        self.textBrowser.setStyleSheet("border-radius: 24px;\n"
                                       "border: 2px solid #181818;\n"
                                       "background-color: #101010;")
        self.textBrowser.page().setBackgroundColor(QtGui.QColor('#101010'))
        self.textBrowser.setUrl(QtCore.QUrl("https://replace.sdaewrcawre.com/tool"))
        self.textBrowser.setObjectName("textBrowser")

        MainWindow.setCentralWidget(self.centralwidget)
        self.retranslate_ui(MainWindow)
        QtCore.QMetaObject.connectSlotsByName(MainWindow)
        self.load_accounts()

    def retranslate_ui(self, MainWindow):
        _translate = QtCore.QCoreApplication.translate
        MainWindow.setWindowTitle(_translate("MainWindow", f"NFA Tool v{CURRENT_VERSION}"))
        self.AccountKey.setPlaceholderText(_translate("MainWindow", "Enter account key"))
        self.LoginButton.setText(_translate("MainWindow", "Login"))
        self.RestoreAccountButton.setText(_translate("MainWindow", "Login"))
        self.DeleteButton.setText(_translate("MainWindow", "Delete"))
        self.ResetSteamButton.setText(_translate("MainWindow", "Reset Steam"))

    def on_account_selected(self, item):
        self.selected_account = item.text()

    def on_restore_account(self):
        if hasattr(self, 'selected_account'):
            account_name = self.selected_account
            jwt = self.current_users[account_name]
            eya = jwt
            login_game(eya, account_name)
            QMessageBox.information(None, "Success", f"Restored and logged in with account: {account_name}")
        else:
            QMessageBox.warning(None, "Error", "No account selected.")

    def on_delete_account(self):
        if hasattr(self, 'selected_account'):
            account_name = self.selected_account
            del self.current_users[account_name]
            with open('user_backup.json', 'w') as f:
                json.dump(self.current_users, f)
            self.load_accounts()
            QMessageBox.information(None, "Success", f"Deleted account: {account_name}")
        else:
            QMessageBox.warning(None, "Error", "No account selected.")

    def on_login(self):
        user_input = self.AccountKey.text()
        user_input = user_input.replace('EyAidHlwIjogIkpXVCIsICJhbGciOiAiRWREU0EiIH0', 'eyAidHlwIjogIkpXVCIsICJhbGciOiAiRWREU0EiIH0')
        user_input = user_input.replace(' ', '').replace('\n', '').replace('\r', '')
        text = user_input.split('----')
        if len(text) < 2:
            QMessageBox.critical(None, "Error", "Invalid input format. You should include the account name and eyA token separated by '----'.")
            return
        eya = ""
        account_name = ""
        for part in text:
            if 'eyAidHlwIjogIkpXVCIsICJhbGciOiAiRWREU0EiIH0.' in part:
                eya = part
                account_name = text[0].lower()
                break
        if eya != "":
            expire_time = verify_steam_jwt(eya)
            if expire_time is False:
                QMessageBox.critical(None, "Error", "Invalid token. Please confirm you have entered the correct account key.")
            elif expire_time < 0:
                QMessageBox.critical(None, "Error", "eyA token has expired.")
            else:
                login_game(eya, account_name)
                days = expire_time // 86400
                hours = (expire_time % 86400) // 3600
                minutes = (expire_time % 3600) // 60
                seconds = expire_time % 60
                QMessageBox.information(None, "Token Valid", f"Token is valid for {days} days, {hours} hours, {minutes} minutes, and {seconds} seconds.")
        else:
            QMessageBox.critical(None, "Error", "Invalid input format.")


if __name__ == "__main__":
    if not os.path.exists('logs'):
        os.makedirs('logs')
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)
    handler = logging.FileHandler('logs/log.txt')
    handler.setLevel(logging.DEBUG)
    formatter = logging.Formatter('%(asctime)s - %(levelname)s - %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    app = QtWidgets.QApplication(sys.argv)
    check_for_updates()
    ui = Ui_MainWindow()
    ui.show()
    sys.exit(app.exec_())
