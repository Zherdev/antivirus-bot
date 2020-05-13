Антивирус-бот для ICQ
=====================

Выполнен Жердевым Иваном в рамках конкурса ICQ ботов Mail.ru.
Антивирус-бот проверяет файлы пользователей на наличие угроз. 

Имя бота в ICQ: @antivirus_bot

Используются антивирусы:

+ [ClamAV](https://www.clamav.net/)

Используются технологии:

+ [go](https://golang.org/)
+ [bot-golang](https://github.com/mail-ru-im/bot-golang) - для работы с API ICQ New
+ [echo](https://github.com/labstack/echo/) - фреймворк для работы с сетью
+ [logrus](https://github.com/sirupsen/logrus) - логирование
+ [docker](https://www.docker.com/) - запуск в контейнерах
+ [aws](https://aws.amazon.com) - работает в облаке Amazon Web Services

Алгоритм работы приложения
--------------------------

AntivirusBot обрабатывает запросы пользователей параллельно.
FileChecker скачивает файлы на проверку в порядке очереди,
проверку в антивирусах осуществляет параллельно. 
Логи сохраняются в директории согласно config.json. 

![flow_start image](assets/flow_start.svg)

![flow_file image](assets/flow_file.svg)

![flow_url image](assets/flow_url.svg)

Запуск и администрирование приложения
-------------------------------------

Склонируйте репозиторий, в корне репозитория разместите конфигурационный файл config.json:

```
{
    "IcqToken": "ваш icq api токен",
   
    "FileBufferSize" : максимальный_размер_очереди (100),
    "FileMaxSize": максимальный_размер_файла (50000000),
    "DownloadTimeout": максимальное_время_загрузки_файла (30),
   
    Путь к логам:
    "BotLogDir": "logs/bot",
    "CheckerLogDir": "logs/checker",
    "AdminLogDir": "logs/admin",
    "FilesDir": "files", <- сюда будут скачиваться файлы на проверку
   
    Настройка администрирования:
    "AdminHost": "0.0.0.0", хост на котором будет запущен бот
    "AdminShutdownPath": "/admin/shutdown",
    "AdminGetLogsPath": "/admin/getLogs/",
    "AdminLogsPath": "logs/", относительный путь к папке, из которой админка берет логи
    "AdminPort": "8085", порт должен совпадать с тем, который открывается в docker-compose.yml
    "AdminToken": "ваш собственный админский токен",
    "AdminTimeout": таймаут_для_админки
}
```

Для запуска выполните в директории deployments/docker/ команду: 

```docker-compose up --build --abort-on-container-exit```

Для администрирования выполните соответствующий POST запрос с телом:
```
{
    "AdminToken": "ваш собственный админский токен"
}
```
по адресу, который установлен в config.json:
+ /admin/shutdown - остановка приложения
+ /admin/getLogs/... - получить логи, например, /admin/getLogs/bot/1589214916

Расширение приложения
---------------------

Вы можете легко добавить поддержку нового антивируса в бот. 
Для этого реализуйте интерфейс antivirusClients.Client:
```
type Client interface {
	CheckFile(fileForCheck *common.FileForCheck, checkResult chan *common.AntivirusResult)
}
```

Здесь единственный метод CheckFile предназначен для проверки файла в антивирусе.
Реализацию следует разместить в пакете antivirusClients.

Клиенты-антивирусы добавляются в бот при инициализации приложения в application.NewApp. 
Пример:
```
myAntivirusClient := antivirusClients.NewMyAntivirusClient()
_ = a.checker.AddAv(myAntivirusClient)
```
Теперь каждый файл будет проходить проверку через соответствующий антивирус.
Сам антивирус лучше запускать в контейнере, внеся соответствующие изменения в docker-compose.yml.

История версий
--------------
* v2.2. Добавлена возможность получить логи из админки по http
* v2.1. Улучшена проверка пользовательских ссылок
* v2. Добавлена проверка файлов по ссылкам. Изменен интерфейс antivirusClients.Client
* v1.1 hotfix. Исправлены ошибки
* v1. Первый релиз
