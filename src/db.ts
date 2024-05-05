import {PrismaClient} from "@prisma/client"

export const DB = new PrismaClient();
export async function handleDBShutdown() {
    await DB.$disconnect();
}