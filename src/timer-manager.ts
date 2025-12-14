import type {Timer} from "@prisma/client";
import {EmbedBuilder, type TextBasedChannel, type TextChannel, User} from "discord.js";
import {DB} from "./db.ts";
import * as crypto from "node:crypto";
import {client} from "./client.ts";
import {formatDistance} from "date-fns";

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

export async function editTimer(timer: Timer, newMessage: string | null, newDue: Date | null): Promise<Timer> {
    if (!newMessage && !newDue) {
        return timer;
    }

    const updatedMessage = newMessage ?? timer.message;
    const updatedDue = newDue ?? timer.due;
    const updatedSnoozedDue = newDue ?? timer.snoozedDue;

    return DB.timer.update({
        where: {
            id: timer.id,
        },
        data: {
            message: updatedMessage,
            due: updatedDue,
            snoozedDue: updatedSnoozedDue,
        },
    });
}

async function createTimerId(): Promise<string> {
    const internalCreateTimerId = () => {
        const bytes = crypto.randomBytes(4);
        const letters = "abcdefghijklmnopqrstuvwxyz";
        let id = "";
        for (let i = 0; i < 4; i++) {
            id += letters[bytes[i] % letters.length];
        }
        return id;
    };

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

    const embed = createTimerEmbed(timer, TimerEmbedInfo.TIMER_DUE, owner);

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
    TIMER_EDITED: {title: "Timer edited", color: 0xffff00},
    TIMER_DUE: {title: "Timer due", color: 0x0000ff}
};

function formatDate(date: Date, includeDuration: boolean): string {
    let base = date.toLocaleString("en-GB");

    if (includeDuration) {
        base += "\n";
        base += `(${formatDistance(date, new Date(), { addSuffix: true })})`
    }

    return base;
}

export function createTimerEmbed(timer: Timer, embedInfo: TimerEmbedInfo, owner: User): EmbedBuilder {
    let due: string;
    let created: string;
    switch (embedInfo) {
        case TimerEmbedInfo.TIMER_CREATED:
            due = formatDate(timer.snoozedDue, true)
            created = formatDate(timer.creation, false)
            break;

        case TimerEmbedInfo.TIMER_DELETED:
            due = formatDate(timer.snoozedDue, true)
            created = formatDate(timer.creation, true)
            break;

        case TimerEmbedInfo.TIMER_SNOOZED:
            due = formatDate(timer.snoozedDue, true)
            created = formatDate(timer.creation, true)
            break;

        case TimerEmbedInfo.TIMER_EDITED:
            due = formatDate(timer.snoozedDue, true)
            created = formatDate(timer.creation, true)
            break;

        case TimerEmbedInfo.TIMER_DUE:
            due = formatDate(timer.snoozedDue, false)
            created = formatDate(timer.creation, true)
            break;

        default:
            due = formatDate(timer.snoozedDue, false)
            created = formatDate(timer.creation, false)
    }

    const embed = new EmbedBuilder()
        .setColor(embedInfo.color)
        .setTitle(embedInfo.title)
        .setDescription(timer.message)
        .addFields(
            {name: "Id", value: timer.id, inline: true},
            {name: "Owner", value: `<@${owner.id}>`, inline: true},
            {name: "\u200b", value: "\u200b", inline: true},
            {name: "Due", value: due, inline: true},
            {name: "Created", value: created, inline: true}
        );

    if (timer.snoozeCount > 0) {
        embed.addFields({name: "Snooze count", value: timer.snoozeCount.toString(), inline: true});
    }

    return embed;
}