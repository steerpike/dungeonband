This project uses beads `bd quickstart` to manage work and progress.

It should use Honeycomb to send opentelemetry traces and development should be observability focused to allow an eventual adoption of "debug in production". This means that feature development isn't complete until it is measurable within Honeycomb.

To faciliate this Opencode should configure the available Honeycomb MCP server for availbility whenever Opencode is started.
Details for configuring the Honeycomb MCP are located at:
https://docs.honeycomb.io/integrations/mcp/configuration-guide/
Details for configuring an MCP server for Opencode are at:
https://opencode.ai/docs/mcp-servers/

There is a HONEYCOMB_MCP variable that is exported as part of my ~/.zshrc and has the key:value to connect to the appropriate datasources in Honeycomb and use the available MCP tools.

There is a HONEYCOMB_DUNGEONBAND_API_KEY variable, a HONEYCOMB_DUNGEONBAND_API_KEY_ID variable and a HONEYCOMB_DUNGEONBAND_DATASET variable set in a .env file in this directory which should be used as needed to actually send the project telemetry to Honeycomb.

I want to create a party based roguelike similar to angband, but where the user control a party of characters of different classes rather than a single character.

It should utilise symbols and characters to represent the map and the created creatures and party.

It should ideally use golang as the development language.

Core game loop should consist of exploring where the party is aggregated into a single `&` symbol with some aggretated stats for moving around the dungeon (stealth, perception, intimidation, potentially others) which influence dungeon encounters. Once the party is spotted by monsters or the party engages with monsters the game shifts to tactical, turn based combat mode where each party member is "unpacked" into the surrounding room as best possible given the available squares and the enemies are similarily "unpacked" from their initial state (although monsters can be singular or grouped when encountered in explore mode).

I would like to start with basic map generation with `#` for walls and `.` for floor.

Then create a party entity and then we can work on shifting state between explore and combat modes.