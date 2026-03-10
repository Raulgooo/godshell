import { loadFont as loadInter } from "@remotion/google-fonts/Inter";
import { loadFont as loadJetBrainsMono } from "@remotion/google-fonts/JetBrainsMono";

export const { fontFamily: interFamily } = loadInter("normal", {
  weights: ["200", "300", "400", "600", "700", "800"],
  subsets: ["latin"],
});

export const { fontFamily: jetbrainsFamily } = loadJetBrainsMono("normal", {
  weights: ["400", "600", "700"],
  subsets: ["latin"],
});
