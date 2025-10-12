# Example Movement Snacks (Movos)

This directory contains example YAML files to demonstrate the movodoro format.

## Using These Examples

**Option 1: Copy and customize**
```bash
mkdir ~/my-movos
cp movos-examples/* ~/my-movos/
export MOVODORO_MOVOS_DIR=~/my-movos
```

**Option 2: Create your own from scratch**

See the main README.md for full documentation on the YAML format.

## What's Included

- **breathing.yaml** - Simple breathing exercises (RPE 1)
- **mobility.yaml** - Joint mobility and stretching (RPE 2-3)
- **bodyweight-strength.yaml** - Bodyweight strength work (RPE 4-6)

## Customizing for Your Needs

These are just starting points. You should:
1. Adjust RPE values to match YOUR perception of difficulty
2. Add your own equipment (kettlebells, clubs, bands, etc.)
3. Create categories that match your training style
4. Set `every_day: true` for movements you want to prioritize
5. Add tags that help you filter (equipment, body regions, etc.)

## Tag Conventions

All tags should end with 'x' for easy grepping:
- Equipment: `kbx`, `clubsx`, `bandsx`
- Body: `bodyx` (bodyweight only)
- Type: `breathx`, `strengthx`, `mobilityx`, `corex`, `flowx`

Happy moving! 🏃‍♂️
