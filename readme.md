Setup.exe
=========

Search `*.msi` from setup.ini and call it.

```
[Product0]
MsiPath1=Installer64\hogehoge.msi
```

When the previous version of `*.msi` is installed,

```
msiexec /i Installer64\hogehoge.msi REINSTALL=ALL REINSTALLMODE=vomus
```

else

```
msiexec /i Installer64\hogehoge.msi
```
