Tages — Golang Test Task 


Требования

1. Принимать бинарные файлы (изображения) от клиента и сохранять их на жёсткий диск
2. Возможность просмотра списка всех загруженных файлов: Имя файла · Дата создания · Дата обновления 
3. Ограничивать количество одновременных подключений к загрузке/скачиванию файлов — **10** 
4. Ограничивать количество одновременных подключений к просмотру списка файлов — **100**

## Запуск

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
```

## Тестирование в Postman

<img width="831" height="501" alt="2025-11-21_02-09-29" src="https://github.com/user-attachments/assets/3a8a772c-3057-4e77-a7e8-03f96e65d9d2" />
<img width="831" height="501" alt="2025-11-21_02-09-29" src="https://github.com/user-attachments/assets/7f52f85e-a3bd-497c-b51d-b177251e33eb" />
<img width="831" height="605" alt="2025-11-21_02-10-11" src="https://github.com/user-attachments/assets/2325cc75-4f72-4265-831f-ee9e9d490e3e" />
