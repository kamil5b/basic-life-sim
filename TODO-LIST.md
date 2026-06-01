## Room Management

### Sell / Trash Item
Right now there is no option to sell / trash any items

## Food Management

### Buying food management
Steps:
- Buy food
- Place anywhere on the world, it will take up 1 Z (the lowest it can get, if there is no item then 0, if there is item, stack it)
- If place on the fridge, it will be stored in the fridge, and it will take up space in the fridge, and user have to organize the food inside the fridge like you organize placable item inside your room. The dimension of the fridge is defined in entity.
- Food in fridge will spoil longer
- Food in fridge in the cold area will spoil even longer

### Food Logistics
- Food that inside fridge can be take out and have to be placed somewhere else, it will spoil faster than if it is in the fridge, but it will still spoil slower than food that is not in the fridge (minimum, today)
- Food that is not in the fridge can be moved around
- If food placed on top of another food, it will stack and become a different item. But with permission, user can choose wether to mix it or not, if not then user will have to put it somewhere else, if yes then it will become a different item.

### Cooking logic
- Add new buyable item category: utilities. It include knife, pan, wok, etc.
- If food placed on the kitchen utilities, it will change form (if the utilities have abilities like knife, chopper, grinder, etc.) OR it can be placed on the utilities (if the utilities is like wok, pan, etc.)
- Cooking utilities (like pan, wok, etc) can be placed on the stove / oven / etc (that can cook things), and it will cook the food inside it.

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
