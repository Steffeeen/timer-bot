import {Events, Routes} from "discord.js";
import {client} from "./client.ts";
import {handleDBShutdown} from "./db.ts";
import {Command, TimerCommand, UntilCommand} from "./command.ts";

const token = process.env["TOKEN"];
const applicationId = process.env["APPLICATION_ID"];

if (!token || !applicationId) {
    console.error("No token or application ID provided");
    process.exit(1);
}

const commands: Command[] = [
    new UntilCommand(),
    new TimerCommand(),
];

// noinspection JSIgnoredPromiseFromCall
client.login(token);

registerCommands(commands, applicationId);

client.on(Events.InteractionCreate, async interaction => {
    if (interaction.isChatInputCommand()) {
        const command = commands.find(command => command.data.name === interaction.commandName);
        if (!command) return;

        await command.execute(interaction);
    } else if (interaction.isAutocomplete()) {
        const command = commands.find(command => command.data.name === interaction.commandName);
        if (!command) return;

        await command.autocomplete(interaction);
    }
});

function registerCommands(commands: Command[], applicationId: string) {
    const data = commands.map(command => command.data.toJSON());
    client.rest.put(Routes.applicationCommands(applicationId), {body: data});
}

async function handleShutdown() {
    await handleDBShutdown();
    process.exit(0);
}

process.on("exit", handleShutdown);
process.on("SIGTERM", handleShutdown);
process.on("SIGINT", handleShutdown);
