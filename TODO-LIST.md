## UI

### Remove Main Tabs
- Home: Map of our room, and the items inside it. User can click on the item to see its details and the stacks (if its a stacks). If choosing fridge, it will show the food inside the fridge and its detail, and can be taken out.
- User click door to go to Shop (it can be to uni / work in future).
- Shop: List of items that can be bought, including food, utilities, placeable items. User can click on the item to see its details and buy it. User can also search for items by name or category.

## Gameplay

### Time management
- Time have to be moving and it can be controlled by the user. Default: 1 second = 1 minute, but user can change it to 1 second = 2 minutes, or 1 second = 0.5 minute, etc. Time will affect the food spoilage, and the time of day (day / night).
- Some activity can skip time, like sleeping, watching TV, etc. User can choose how much time to skip, and it will affect the food spoilage, and the time of day (day / night).

### Save / Load
- User can save the game at any time, and load it later. It will save the state of the world, including the items, food, time, etc. User can have multiple save files, and load any of them. User can also delete save files.
- Use go/gob to save / load the game, and it will be faster than using file system. It will also allow us to save / load the game in the cloud, and share it with other users.
