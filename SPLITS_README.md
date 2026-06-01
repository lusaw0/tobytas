This is a detailed explanation of everything contained within splits.json.
This document is basically a fuck-you to whoever decided comments are forbidden in JSON.

Below is what the json file would look like if comments were supported:

```json
{
    "game_name": "Undertale", // Game name, don't change unless the game isn't UT.
    "category_name": "New Game+ Neutral", // Category, I have it set to NG+ Neutral but change it at will.
    "pb_attempts": 38, // Finished runs, dogshit variable name
    "total_attempts": 1420, // Total attempts
    "start_frame": 76, // Frame of the movie file it starts at (76f @ constant 60fps = 1.266s)
    "split_frames": [
        19397, // Ruins split, Flowey room door touch [frame the alarm starts counting]
        37820, // Snowdin split, frame before you enter Snowdin Town
        49152, // Papyrus split, frame before you enter room_water1
        69114, // Spears 2 split, frame that the screen cuts to black
        85200, // Enter Lab split, frame before you enter Alphys' lab
        91844, // RF2 split, frame before you enter RF2 for the first time from the elevator
        96996, // LF3 split, frame before you enter LF3 for the first time from the elevator
        103075, // RF3 split, frame before you enter RF3 for the first time from the elevator
        108323, // Long Elevator split, frame before you enter the room just after elevator exit.
        115400, // New Home split, frame before you enter barrier room.
        125184  // End split, post-barrier door touch [frame the alarm starts counting]
    ],
    "splits": [
        //      name = the name of the split.
        //   pb_time = the pb time in nanoseconds
        // best_time = the gold in nanoseconds.
        {"name": "Ruins",         "pb_time": 406490000000, "best_time": 390166079300},
        {"name": "Snowdin",       "pb_time": 756080000000, "best_time": 335476656100},
        {"name": "Papyrus",       "pb_time": 954596000000, "best_time": 191578738100},
        {"name": "Spears 2",      "pb_time": 1303269000000, "best_time": 348673000000},
        {"name": "Enter Lab",     "pb_time": 1588958000000, "best_time": 284962058600},
        {"name": "Right Floor 2", "pb_time": 1708788000000, "best_time": 116035542800},
        {"name": "Left Floor 3",  "pb_time": 1800423000000, "best_time": 91217191100},
        {"name": "Right Floor 3", "pb_time": 1911720000000, "best_time": 105919000000},
        {"name": "Long Elevator", "pb_time": 2011314000000, "best_time": 93871747500},
        {"name": "New Home",      "pb_time": 2250217000000, "best_time": 238903000000},
        {"name": "End",           "pb_time": 3056429000000, "best_time": 657811986800}
    ]
}
```