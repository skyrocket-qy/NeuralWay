# 2D Pixel Tower Defense

A Diablo-like TD(Tower Defense) game.

## Possible Features
- Multi player(cooperate, against)
- Workshop

## Steps
1. Select the TD map
2. deploy the heroes to the place
3. start the game

## Game mechanism
- Monster will keep find the shortest path to end
- if in the hero range, hero will attack monster
- the end has life
- if monster reach the end, the life will reduce - 1
- if end life == 0, game over
- if all monsters killed, win

## Recent Updates & New Mechanisms

### 1. Player Data Loading from JSON
The `Player.cs` script now loads hero data from a JSON file instead of a plain text file.
*   **File:** `Assets/Scripts/Player.cs`
*   **Setup:**
    1.  Create a `hero_data.json` file in your `Resources` folder (e.g., `Assets/Resources/hero_data.json`).
    2.  Populate it with hero names in the following JSON format:
        ```json
        {
            "heroNames": [
                "Hero1Map",
                "Hero2Map",
                "AnotherHeroMap"
            ]
        }
        ```

### 2. Rogue-like Card System (Data Structure)
A new ScriptableObject `CardData` has been introduced to define the properties of cards that can be offered to the player.
*   **File:** `Assets/Scripts/Progression/CardData.cs`
*   **Usage:**
    1.  In Unity, go to `Assets -> Create -> Game -> Card Data` to create new card assets.
    2.  Fill in the `cardName`, `description`, `icon`, `effectType`, and `effectValue` for each card.
    3.  The `ApplyEffect` method in `CardData.cs` is a placeholder for implementing the card's actual effect on the `Player` or `Hero`.

### 3. Hero Experience & Leveling System
Heroes now have an experience and leveling progression system.
*   **File:** `Assets/Scripts/Hero.cs`
*   **Details:**
    *   Each `Hero` instance tracks `currentExperience`, `experienceToNextLevel`, and `level`.
    *   The `GainExperience(int amount)` method handles experience accumulation and triggers `LevelUp()` when enough experience is gained.
    *   `LevelUp()` increases the hero's level, adjusts `currentExperience`, and recalculates `experienceToNextLevel`.

### 4. Wave-based Experience Distribution
The `GameManager` now tracks monster deaths and awards experience to active heroes upon wave completion.
*   **Files:** `Assets/Scripts/Enemies/Monster.cs`, `Assets/Scripts/GameManager.cs`
*   **How it works:**
    1.  `Monster.cs` now includes a `public static event Action<Monster> OnMonsterDied;` which is invoked when a monster is defeated.
    2.  `GameManager.cs` subscribes to this event.
    3.  `GameManager` maintains a list of `activeMonsters` for the current wave.
    4.  When `GenMonster()` is called, the new monster is added to `activeMonsters`.
    5.  When `Monster.OnMonsterDied` is triggered, `GameManager` removes the monster from `activeMonsters`.
    6.  The `StartWaves()` coroutine now waits until all monsters for the current wave have been *generated* AND all `activeMonsters` have been *defeated*.
    7.  Once a wave is cleared, `GameManager` calls `AwardExperienceToHeroes()`, which finds all active `Hero` components and calls their `GainExperience()` method with the `currentWaveExperience` (a public variable in `GameManager` that can be set in the Inspector).