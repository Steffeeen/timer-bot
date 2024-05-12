FROM node:17

WORKDIR /app

COPY prisma/schema.prisma /app/prisma/schema.prisma
COPY src /app/src
COPY package.json /app
COPY package-lock.json /app
COPY tsconfig.json /app

RUN npm install

ENV DATABASE_URL="file:../data/timerbot.db"

RUN npx prisma migrate dev --name init

CMD npx prisma migrate dev --name init && npm run start
