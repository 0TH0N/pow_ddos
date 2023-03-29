## PET проект по тестовому заданию по работе с алгоритмом POW для защиты от DDOS атак ##

*Текст задачи в конце readme.*

Для работы "proof of work" в данном проекте был выбран алгоритм SHA256
как широко зарекомендовавший себя и имеющий репутацию надежного, в том 
числе благодаря использованию его в основе биткоина.

Проект можно запустить командой *"docker-compose up --build"* в корне проекта.

При запуске стартует сервер и эмулятор работы клиента, который раз в 2
секунды отправляет запрос на сервер и выводит информацию о ходе обработки
запросов, в том числе предоставляя информацию о количестве затраченного
времени на некоторые этапы.

Также для запросов "руками" в проекте есть файл *"api_requests.http"* 
в котором содержатся подготовленные образцы запросов для встроенного 
http клиента для IDE от Jetbrains.

**Логика работы сервера следующая:**

1. Сервер получает запрос без указания заголовков *"Pow_task"* и *"Pow_hash"*
и отправляет ответ без тела с заголовком *"Pow_task"*, который необходимо 
"решить" клиенту. Также запоминает на 5 минут переданную задачу как 
"нерешенную".
2. Клиент на своей стороне вычисляет из полученного *"Pow_task"* нужный хэш и
отправляет новый запрос с пустым телом и заголовками *"Pow_task"* и 
*"Pow_hash"*.
3. Сервер проверяет переданную задачу на наличие в своем кэше, "нерешенный"
статус и проверяет правильность хэша. В случае успеха возвращает
произвольную фразу из "Wisdom quotes" и помечает задачу как "решенную".

Примечание: Запоминание на короткий срок выданных и решенных задач позволит 
избежать атаки с предварительной подготовкой большого количества пар данных 
"ключ - верный хэш".



>Test task for Server Engineer
Design and implement “Word of Wisdom” tcp server.
• TCP server should be protected from DDOS attacks with the Prof of Work
(https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should
be used.
• The choice of the POW algorithm should be explained.
• After Prof Of Work verification, server should send one of the quotes from “word of
wisdom” book or any other collection of the quotes.
• Docker file should be provided both for the server and for the client that solves the
POW challenge.