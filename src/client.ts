import { Client, Events, GatewayIntentBits } from "discord.js";

export const client = new Client({ intents: [GatewayIntentBits.Guilds] });

client.once(Events.ClientReady, client => {
    console.log(`Logged in as ${client.user.tag}`);
});
