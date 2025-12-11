# Q-Flow Frontend (Vue 3 + Vite)

Минимальный SPA-клиент для бэкенда Q-Flow.

## Установка и запуск
```bash
cd q-flow-frontend
npm install
npm run dev    # http://localhost:5173
npm run build  # прод сборка в dist
```

## Конфигурация
- Базовый URL API читается из `localStorage.qflow_api` или `http://localhost:8080`. Его можно изменить в UI (поле API base) или вручную в `localStorage`.
- JWT хранится в `localStorage.qflow_token`.

## Функции
- Регистрация/логин (email, ФИО, пароль).
- Привязка Telegram (username, chat_id) и получение ссылки `/start`.
- Каталог очередей по коду группы, создание, редактирование (title/description), архивирование и удаление.
- Операции с очередью: вступление (slots поддерживают slot_time), выход, продвижение, ручное добавление участника владельцем, удаление участника.
- Просмотр участников и их позиций.
