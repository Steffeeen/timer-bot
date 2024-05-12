FROM node:17

WORKDIR /app

COPY prisma/schema.prisma /app/prisma/schema.prisma
COPY src /app/src
COPY package.json /app
COPY package-lock.json /app
COPY tsconfig.json /app

RUN npm install


ENV DB_FILE="/app/data/timerbot.db"
ENV DATABASE_URL="file:${DB_FILE}"
ENV TZ="Europe/Berlin"

RUN npx prisma migrate dev --name init

RUN echo "if [ ! -f "$DB_FILE" ]; then" > setup-db.sh
RUN echo "  echo "SQLite database not found. Running prisma migrate deploy..."" >> setup-db.sh
RUN echo "  npx prisma migrate deploy" >> setup-db.sh
RUN echo "else" >> setup-db.sh
RUN echo "  echo "SQLite database already exists."" >> setup-db.sh
RUN echo "fi" >> setup-db.sh

RUN chmod +x setup-db.sh

CMD bash setup-db.sh && npm run start
