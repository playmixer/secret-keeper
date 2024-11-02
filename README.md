# Менеджер паролей GophKeeper
## Общие требования
GophKeeper представляет собой клиент-серверную систему, позволяющую пользователю надёжно и безопасно хранить логины, пароли, бинарные данные и прочую приватную информацию.


### Сервер должен реализовывать следующую бизнес-логику:
* регистрация, аутентификация и авторизация пользователей;
* хранение приватных данных;
* синхронизация данных между несколькими авторизованными клиентами одного владельца;
* передача приватных данных владельцу по запросу.
### Клиент должен реализовывать следующую бизнес-логику:
* аутентификация и авторизация пользователей на удалённом сервере;
* доступ к приватным данным по запросу.
### Функции, реализация которых остаётся на усмотрение исполнителя:
* создание, редактирование и удаление данных на стороне сервера или клиента;
* формат регистрации нового пользователя;
* выбор хранилища и формат хранения данных;
* обеспечение безопасности передачи и хранения данных;
* протокол взаимодействия клиента и сервера;
* механизмы аутентификации пользователя и авторизации доступа к информации.
### Дополнительные требования:
* клиент должен распространяться в виде CLI-приложения с возможностью запуска на платформах Windows, Linux и Mac OS;
* клиент должен давать пользователю возможность получить информацию о версии и дате сборки бинарного файла клиента.
## Типы хранимой информации
* пары логин/пароль;
* произвольные текстовые данные;
* произвольные бинарные данные;
* данные банковских карт.
Для любых данных должна быть возможность хранения произвольной текстовой метаинформации (принадлежность данных к веб-сайту, личности или банку, списки одноразовых кодов активации и прочее).
## Абстрактная схема взаимодействия с системой
Ниже описаны базовые сценарии взаимодействия пользователя с системой. Они не являются исчерпывающими — решение отдельных сценариев (например, разрешение конфликтов данных на сервере) остаётся на усмотрение исполнителя.
### Для нового пользователя:
1. Пользователь получает клиент под необходимую ему платформу.
2. Пользователь проходит процедуру первичной регистрации.
3. Пользователь добавляет в клиент новые данные.
4. Клиент синхронизирует данные с сервером.
### Для существующего пользователя:
1. Пользователь получает клиент под необходимую ему платформу.
2. Пользователь проходит процедуру аутентификации.
3. Клиент синхронизирует данные с сервером.
4. Пользователь запрашивает данные.
5. Клиент отображает данные для пользователя.
## Тестирование и документация
Код всей системы должен быть покрыт юнит-тестами не менее чем на 80%. Каждая экспортированная функция, тип, переменная, а также пакет системы должны содержать исчерпывающую документацию.
## Необязательные функции
Перечисленные ниже функции необязательны к имплементации, однако позволяют лучше оценить степень экспертизы исполнителя. Исполнитель может реализовать любое количество из представленных ниже функций на свой выбор:
* поддержка данных типа OTP (one time password);
* поддержка терминального интерфейса (TUI — terminal user interface);
* использование бинарного протокола;
* наличие функциональных и/или интеграционных тестов;
* описание протокола взаимодействия клиента и сервера в формате Swagger.



# Server GophKeeper
## Run server

### 1. клонируем проект
```env
git clone https://github.com/playmixer/secret-keeper.git
```
### 2. генерируем сертификаты (выполнить из корня проекта) или запустить ./scripts/get_cert.sh
```env
mkdir ./cert
openssl genrsa -out ./cert/gophkeeper.key 2048
openssl ecparam -genkey -name secp384r1 -out ./cert/gophkeeper.key
openssl req -new -x509 -sha256 -key ./cert/gophkeeper.key -out ./cert/gophkeeper.crt -days 3650
```
### 3. запускаем сервер
```bash
cd deploy
docker-compose up
```
### generate swag and run
```bash
swag init -o ./docs -g ./internal/adapter/api/rest/rest.go
swag fmt
go run ./cmd/server/server.go
```
ссылка на swagger
https://localhost:8443/swagger/index.html

## Пример .env
```env
REST_ADDRESS=localhost:8443
SSL_ENABLE=1
LOG_LEVEL=debug
LOG_PATH=./logs/server.log
SECRET_KEY=secret_key
DATABASE_STRING=postgres://root:root@localhost:5432/keeper?sslmode=disable
ENCRYPT_KEY=RZLMAOIOuljexYLh5S47O9kfVI7O1Ll0
```
для ENCRYPT_KEY длина должна быть 32 символа

# Client GophKeeper
## Запуск клиента
#### Вариант 1
* билдим
```bash
go build -ldflags "-X main.buildVersion=1.0.0 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git show --oneline -s)'" ./cmd/client/client.go
```
или ./scripts/build_client.sh
* запускаем скомпилированый файл, для windows ***client.exe***

### Вариант 2
```bash
go run -ldflags "-X main.buildVersion=1.0.0 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git show --oneline -s)'" ./cmd/client/client.go
```
или ./scripts/run_client.sh

```text
По умолчанию клиент подключается к https://localhost:8443
```

### Глобальные переменные клиента
При необходимости рядом с клиентом разместить файл ***.env.client***, 
с содержимым
```env
API_ADDRESS=https://localhost:8443
LOG_LEVEL=debug
LOG_PATH=./logs/client.log
FILE_MAX_SIZE=819200
```

### Тесты
в работе
### покрытие
```shel
go test -v -coverpkg=./... -coverprofile=profile.cov ./...
go tool cover -func profile.cov
```



### fix fieldalignment
```
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
```
```
fieldalignment -fix <package_path>
```


#### Generate private key (.key)
```
# Key considerations for algorithm "RSA" ≥ 2048-bit
openssl genrsa -out server.key 2048

# Key considerations for algorithm "ECDSA" ≥ secp384r1
# List ECDSA the supported curves (openssl ecparam -list_curves)
openssl ecparam -genkey -name secp384r1 -out server.key
```
#### Generation of self-signed(x509) public key (PEM-encodings .pem|.crt) based on the private (.key)
```
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```