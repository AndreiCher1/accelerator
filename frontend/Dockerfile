# Этап 1: Сборка приложения с помощью pnpm
FROM node:20-alpine AS builder

# Установка pnpm глобально
RUN npm install -g pnpm

WORKDIR /app

# Сначала копируем файлы зависимостей для эффективного кэширования слоев Docker
COPY package.json pnpm-lock.yaml* ./

# Устанавливаем зависимости
RUN pnpm install --frozen-lockfile

# Копируем весь оставшийся код фронтенда
COPY . .

# Собираем продакшн-версию проекта (Vite по умолчанию собирает в папку 'dist')
RUN pnpm run build

# Этап 2: Запуск веб-сервера Nginx для раздачи статики
FROM nginx:alpine

# Удаляем стандартный конфиг Nginx
RUN rm /etc/nginx/conf.d/default.conf

# Копируем наш собственный файл конфигурации Nginx
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Копируем собранные файлы из предыдущего этапа сборки в папку Nginx
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]