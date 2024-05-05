import type {Timer} from "@prisma/client";
import {EmbedBuilder, type TextBasedChannel, type TextChannel, User} from "discord.js";
import {DB} from "./db.ts";
import * as crypto from "node:crypto";
import {client} from "./client.ts";
import {formatDistance} from "date-fns";

const TIMER_ID_LENGTH = 3;

setInterval(checkTimersDue, 60 * 1000);
checkTimersDue();

async function checkTimersDue() {
    const dueTimers = await DB.timer.findMany({
        where: {
            AND: [
                {
                    snoozedDue: {lte: new Date()},
                },
                {
                    shown: false,
                },
            ],
        },
    });

    for (const timer of dueTimers) {
        await showDueTimer(timer);
        await DB.timer.update({
            where: {
                id: timer.id,
            },
            data: {
                shown: true,
            },
        });
    }
}

export async function createTimer(message: string, user: User, channel: TextBasedChannel, due: Date): Promise<Timer> {
    const id = await createTimerId();
    return DB.timer.create({
        data: {
            id: id,
            message: message,
            user: user.id,
            channel: channel.id,
            due: due,
            snoozedDue: due,
            creation: new Date()
        },
    });
}

export async function getAllTimersForUser(user: User, onlyActive: boolean): Promise<Timer[]> {
    return DB.timer.findMany({
        where: {
            user: user.id,
            AND: [
                {
                    shown: !onlyActive,
                }
            ]
        },
    });
}

export async function getTimerById(id: string): Promise<Timer | null> {
    return DB.timer.findUnique({
        where: {
            id: id,
        },
    });
}

export async function deleteTimer(timer: Timer) {
    await DB.timer.delete({
        where: {
            id: timer.id,
        },
    });
}

export async function snoozeTimer(timer: Timer, due: Date): Promise<Timer> {
    return DB.timer.update({
        where: {
            id: timer.id,
        },
        data: {
            snoozedDue: due,
            snoozeCount: {
                increment: 1,
            },
            shown: false,
        },
    });
}

async function createTimerId(): Promise<string> {
    const internalCreateTimerId = () =>
        crypto.randomBytes(3).toString("hex");

    let id: string;
    let timerForId: Timer | null;
    do {
        id = internalCreateTimerId();
        timerForId = await DB.timer.findUnique({
            where: {
                id: id,
            },
        });
    } while (timerForId);

    return id;
}

async function showDueTimer(timer: Timer) {
    const fetchedChannel = await client.channels.fetch(timer.channel, {
        cache: true,
        allowUnknownGuild: true,
    });

    const owner = await client.users.fetch(timer.user);
    if (!owner) {
        console.warn("Could not find owner for timer, not showing timer!");
        return undefined;
    }

    let channel: TextBasedChannel;
    if (fetchedChannel) {
        channel = fetchedChannel as TextChannel;
    } else {
        channel = owner.dmChannel ? owner.dmChannel : await owner.createDM();
    }

    const embed = createTimerEmbed(timer, {title: "Timer due", color: 0x0000ff}, owner);

    channel.send({embeds: [embed], content: `<@${owner.id}>`});
}

export interface TimerEmbedInfo {
    title: string;
    color: number;
}

export const TimerEmbedInfo: Record<string, TimerEmbedInfo> = {
    TIMER_CREATED: {title: "Timer created", color: 0x00ff00},
    TIMER_DELETED: {title: "Timer deleted", color: 0xff0000},
    TIMER_SNOOZED: {title: "Timer snoozed", color: 0x00ffff},
};

export function createTimerEmbed(timer: Timer, {title, color}: TimerEmbedInfo, owner: User): EmbedBuilder {
    const embed = new EmbedBuilder()
        .setColor(color)
        .setTitle(title)
        .setDescription(timer.message)
        .addFields(
            {name: "Id", value: timer.id, inline: true},
            {name: "Owner", value: `<@${owner.id}>`, inline: true},
            {name: "Due", value: timer.snoozedDue.toLocaleString("en"), inline: true},
            {name: "Created", value: timer.creation.toLocaleString("en"), inline: true},
            {name: "Duration", value: formatDistance(timer.snoozedDue, timer.creation, {addSuffix: true}), inline: true},
        );

    if (timer.snoozeCount > 0) {
        embed.addFields({name: "Snooze count", value: timer.snoozeCount.toString(), inline: true});
    }

    return embed;
}