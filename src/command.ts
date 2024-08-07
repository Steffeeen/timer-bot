import {AutocompleteInteraction, ChatInputCommandInteraction, EmbedBuilder, SlashCommandBuilder} from "discord.js";
import * as chrono from "chrono-node";
import {formatDistance} from "date-fns";
import {
    createTimer,
    createTimerEmbed,
    deleteTimer, editTimer,
    getAllTimersForUser,
    getTimerById,
    snoozeTimer, TimerEmbedInfo,
} from "./timer-manager.ts";
import type {Timer} from "@prisma/client";

export abstract class Command {
    abstract data: SlashCommandBuilder

    abstract execute(interaction: ChatInputCommandInteraction): Promise<void>

    async autocomplete(interaction: AutocompleteInteraction): Promise<void> {}
}

export class UntilCommand extends Command {
    data: SlashCommandBuilder = new SlashCommandBuilder()
        .setName("until")
        .setDescription("Calculate the time until a date")
        .addStringOption(option => option.setName("date").setDescription("The date to calculate the time until").setRequired(true)) as SlashCommandBuilder;

    async execute(interaction: ChatInputCommandInteraction): Promise<void> {
        const timeString = interaction.options.getString("date");
        if (!timeString) {
            await interaction.reply({
                content: "No date provided",
                ephemeral: true
            });
            return;
        }

        const date = parseTimeString(timeString);
        if (date === null) {
            await interaction.reply({
                content: "Invalid date",
                ephemeral: true
            });
            return;
        }

        const now = new Date();
        const diffString = formatDistance(date, now, {addSuffix: true});
        await interaction.reply({
            embeds: [
                new EmbedBuilder()
                    .setColor(0x00ff00)
                    .setDescription(`Time until ${date.toLocaleString("en-GB")}: ${diffString.replace(RegExp("([0-9]+)"), "**$1**")}`)
            ]
        });
    }
}

export class TimerCommand extends Command {
    data: SlashCommandBuilder = new SlashCommandBuilder()
        .setName("timer")
        .setDescription("Manage timers")
        .addSubcommand(createCommand =>
            createCommand.setName("create")
                .setDescription("Create a new timer")
                .addStringOption(option => option.setName("message").setDescription("The message to display when the timer is up").setRequired(true))
                .addStringOption(option => option.setName("time").setDescription("The time to set the timer for").setRequired(true))
        )
        .addSubcommand(deleteCommand =>
            deleteCommand.setName("delete")
                .setDescription("Delete a timer")
                .addStringOption(option => option.setName("id").setDescription("The ID of the timer to delete").setRequired(true).setAutocomplete(true))
        )
        .addSubcommand(snoozeCommand =>
            snoozeCommand.setName("snooze")
                .setDescription("Snooze a timer")
                .addStringOption(option => option.setName("id").setDescription("The ID of the timer to snooze").setRequired(true).setAutocomplete(true))
                .addStringOption(option => option.setName("time").setDescription("The amount of time to snooze the timer for").setRequired(true))
        )
        .addSubcommand(listCommand =>
            listCommand.setName("list")
                .setDescription("List all timers")
        )
        .addSubcommand(editCommand =>
            editCommand.setName("edit")
                .setDescription("Edit a timer")
                .addStringOption(option => option.setName("id").setDescription("The ID of the timer to edit").setRequired(true).setAutocomplete(true))
                .addStringOption(option => option.setName("message").setDescription("The new message for the timer").setRequired(false))
                .addStringOption(option => option.setName("time").setDescription("The new time for the timer").setRequired(false))
        ) as SlashCommandBuilder;

    async execute(interaction: ChatInputCommandInteraction): Promise<void> {
        switch (interaction.options.getSubcommand()) {
            case "create":
                return handleCreateTimer(interaction);
            case "delete":
                return handleDeleteTimer(interaction);
            case "snooze":
                return handleSnoozeTimer(interaction);
            case "list":
                return handleListTimers(interaction);
            case "edit":
                return handleEditTimer(interaction);
        }

        await interaction.reply({content: "Invalid subcommand", ephemeral: true});
    }

    async autocomplete(interaction: AutocompleteInteraction): Promise<void> {
        const subcommand = interaction.options.getSubcommand();
        if (subcommand === "delete" || subcommand === "snooze" || subcommand === "edit") {
            const focusedOption = interaction.options.getFocused(true);
            if (focusedOption.name === "id") {
                const isSnooze = subcommand === "snooze";
                const timers = await getAllTimersForUser(interaction.user, !isSnooze);
                const options = timers.map(timer => ({
                    name: `${timer.id} - ${timer.message}`,
                    value: timer.id
                })).filter(option => option.name.includes(focusedOption.value));

                await interaction.respond(options);
            }
        }
    }
}

async function handleCreateTimer(interaction: ChatInputCommandInteraction) {
    const message = interaction.options.getString("message");
    const timeString = interaction.options.getString("time");

    if (!message || !timeString) {
        await interaction.reply({
            content: "Missing message or time",
            ephemeral: true
        });
        return;
    }

    const date = parseTimeString(timeString);

    if (!interaction.channel) {
        await interaction.reply({
            content: "No channel found",
            ephemeral: true
        });
        return;
    }

    if (!date) {
        await interaction.reply({
            content: "Invalid time",
            ephemeral: true
        });
        return;
    }

    if ((date.getHours() == 0 || date.getHours() == 12) && date.getMinutes() == 0 && date.getSeconds() == 0) {
        // special case: user entered a date without a time, chrono automatically assumes midnight, change to current time
        const now = new Date();
        date.setHours(now.getHours());
        date.setMinutes(now.getMinutes());
        date.setSeconds(now.getSeconds());
    }

    const timer = await createTimer(message, interaction.user, interaction.channel, date);
    await showTimer(interaction, timer, TimerEmbedInfo.TIMER_CREATED);
}

async function handleDeleteTimer(interaction: ChatInputCommandInteraction) {
    const timer = await getTimerFromInteraction(interaction);
    if (!timer) {
        return;
    }

    await deleteTimer(timer);
    await showTimer(interaction, timer, TimerEmbedInfo.TIMER_DELETED);
}

async function handleSnoozeTimer(interaction: ChatInputCommandInteraction) {
    const timer = await getTimerFromInteraction(interaction);
    if (!timer) {
        return;
    }

    const timeString = interaction.options.getString("time");
    if (!timeString) {
        await interaction.reply({
            content: "No time provided",
            ephemeral: true
        });
        return;
    }

    const date = parseTimeString(timeString);
    if (!date) {
        await interaction.reply({
            content: "Invalid time",
            ephemeral: true
        });
        return;
    }

    const updatedTimer = await snoozeTimer(timer, date);
    await showTimer(interaction, updatedTimer, TimerEmbedInfo.TIMER_SNOOZED);
}

async function handleListTimers(interaction: ChatInputCommandInteraction) {
    const timers = await getAllTimersForUser(interaction.user, true);
    await showTimerList(timers, interaction);
}

async function handleEditTimer(interaction: ChatInputCommandInteraction) {
    const timer = await getTimerFromInteraction(interaction);
    if (!timer) {
        return;
    }

    const message = interaction.options.getString("message");
    const timeString = interaction.options.getString("time");

    if (!message && !timeString) {
        await interaction.reply({
            content: "No message or time provided",
            ephemeral: true
        });
        return;
    }

    let date: Date | null = null;
    if (timeString) {
        date = parseTimeString(timeString);
        if (!date) {
            await interaction.reply({
                content: "Invalid time",
                ephemeral: true
            });
            return;
        }
    }

    const updatedTimer = await editTimer(timer, message, date);
    await showTimer(interaction, updatedTimer, TimerEmbedInfo.TIMER_EDITED);
}

async function getTimerFromInteraction(interaction: ChatInputCommandInteraction): Promise<Timer | null> {
    const timerId = interaction.options.getString("id");

    if (!timerId) {
        await interaction.reply({
            content: "No timer ID provided",
            ephemeral: true
        });
        return null;
    }

    const timer = await getTimerById(timerId);
    if (!timer) {
        await interaction.reply({
            content: `No timer with ID ${timerId} exists`,
            ephemeral: true
        });
        return null;
    }

    if (interaction.user.id !== timer.user) {
        await interaction.reply({
            content: "You do not own this timer",
            ephemeral: true
        });
        return null;
    }

    return timer;
}

function parseTimeString(timeString: string): Date | null {
    let date = chrono.en.GB.parseDate(timeString);

    if (date === null) {
        date = chrono.en.GB.parseDate(`in ${timeString}`)
    }

    if (date === null) {
        date = chrono.en.GB.parseDate(`on ${timeString}`)
    }

    if (date === null) {
        date = chrono.parseDate(timeString)
    }

    return date;
}

async function showTimer(interaction: ChatInputCommandInteraction, timer: Timer, timerEmbedInfo: TimerEmbedInfo) {
    const owner = interaction.user;

    const embed = createTimerEmbed(timer, timerEmbedInfo, owner);
    await interaction.reply({embeds: [embed]});
}

async function showTimerList(timers: Timer[], interaction: ChatInputCommandInteraction) {
    let content = `You have ${timers.length} active timers:\n\n`;
    for (const timer of timers) {
        content += `**${timer.id}** - ${timer.message} - Due: ${timer.snoozedDue.toLocaleString("en-GB")}\n`;
    }

    const embed = new EmbedBuilder()
        .setColor(0x3c1984)
        .setTitle("Active timers")
        .setDescription(content);

    await interaction.reply({embeds: [embed]});
}
