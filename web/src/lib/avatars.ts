import {
  Bird,
  Bug,
  Cat,
  Crown,
  Dog,
  Fish,
  Flame,
  Ghost,
  PawPrint,
  Rabbit,
  Rocket,
  Smile,
  Snail,
  Squirrel,
  Turtle,
  Zap,
  type LucideIcon,
} from "lucide-react";

export interface AvatarDef {
  id: string;
  label: string;
  Icon: LucideIcon;
  bg: string;
  fg: string;
}

export const AVATARS: AvatarDef[] = [
  { id: "ghost", label: "Ghost", Icon: Ghost, bg: "#e9e2ff", fg: "#5f43e8" },
  { id: "cat", label: "Cat", Icon: Cat, bg: "#ffe0e6", fg: "#e8405a" },
  { id: "dog", label: "Dog", Icon: Dog, bg: "#ffe8cf", fg: "#d67a14" },
  { id: "bird", label: "Bird", Icon: Bird, bg: "#dff0ff", fg: "#1785de" },
  { id: "fish", label: "Fish", Icon: Fish, bg: "#d4f6ec", fg: "#0b9a62" },
  { id: "rabbit", label: "Rabbit", Icon: Rabbit, bg: "#f2e6ff", fg: "#8a3fe0" },
  { id: "turtle", label: "Turtle", Icon: Turtle, bg: "#dff6df", fg: "#2e9e4f" },
  { id: "squirrel", label: "Squirrel", Icon: Squirrel, bg: "#ffead2", fg: "#c06a1c" },
  { id: "snail", label: "Snail", Icon: Snail, bg: "#fbe6ef", fg: "#e0497f" },
  { id: "bug", label: "Bug", Icon: Bug, bg: "#e2f7f1", fg: "#0e9c74" },
  { id: "paw", label: "Paw", Icon: PawPrint, bg: "#f2e8d8", fg: "#8c6a4f" },
  { id: "rocket", label: "Rocket", Icon: Rocket, bg: "#ffe4db", fg: "#f2501f" },
  { id: "zap", label: "Zap", Icon: Zap, bg: "#fff2ca", fg: "#d99a00" },
  { id: "crown", label: "Crown", Icon: Crown, bg: "#fdebd0", fg: "#e08a1e" },
  { id: "flame", label: "Flame", Icon: Flame, bg: "#ffe6de", fg: "#ef4a2a" },
  { id: "smile", label: "Smile", Icon: Smile, bg: "#e6f3ff", fg: "#2f7fe0" },
];

const BY_ID = new Map(AVATARS.map((a) => [a.id, a]));

export function avatarById(id: string | null | undefined): AvatarDef {
  return BY_ID.get(id ?? "") ?? AVATARS[0];
}
