Tages — Golang Test Task 


Требования
| 1 | Принимать бинарные файлы (изображения) от клиента и сохранять их на жёсткий диск
| 2 | Возможность просмотра списка всех загруженных файлов: Имя файла · Дата создания · Дата обновления 
| 3 | Ограничивать количество одновременных подключений к загрузке/скачиванию файлов — **10** 
| 4 | Ограничивать количество одновременных подключений к просмотру списка файлов — **100**


## Структура проекта

tages-file-service/
├── cmd/
│   ├── server/    # gRPC-сервер
│   └── client/    # CLI-клиент
├── internal/
│   ├── limiter/   # лимитер
│   ├── server/    # реализация методов gRPC
│   └── storage/   # работа с файлами
├── proto/         # file_service.proto + сгенерированные файлы
├── uploads/       # файлы
├── go.mod


## Быстрый запуск

```bash
# 1. Запуск сервера
go run cmd/server/main.go
# → Listening on :50051

# 2. CLI-клиент
go run cmd/client/main.go --action list
go run cmd/client/main.go --action upload --file ./pic.jpg
go run cmd/client/main.go --action download --name pic.jpg

# 3. Или через grpcurl
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext localhost:50051 file_service.FileService/ListFiles

# 4. Или через Postman gRPC (нужно выбрать .proto файл)
