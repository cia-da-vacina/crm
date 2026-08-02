/**
 * Shared motion variants — polished feedback without gamey flair.
 */
export const easeOut = [0.22, 1, 0.36, 1] as const;

/** Apple-like decelerate: fast start, long soft settle. */
export const appleEase = [0.32, 0.72, 0, 1] as const;

export const fadeUp = {
  initial: { opacity: 0, y: 10 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: 6 },
  transition: { duration: 0.28, ease: easeOut },
};

export const staggerContainer = {
  animate: {
    transition: { staggerChildren: 0.035, delayChildren: 0.04 },
  },
};

export const staggerItem = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.24, ease: easeOut },
};

export const pressable = {
  whileTap: { scale: 0.98 },
  transition: { duration: 0.12 },
};

/** Dimmed + blurred scrim behind centered dialogs. */
export const modalBackdrop = {
  initial: { opacity: 0 },
  animate: { opacity: 1 },
  exit: { opacity: 0 },
  transition: { duration: 0.28, ease: appleEase },
};

/**
 * Panel open like a macOS window/alert: slight oversize + soft spring settle,
 * with a short fade so it never pops hard.
 */
export const modalPanel = {
  initial: { opacity: 0, scale: 1.1, y: 12 },
  animate: {
    opacity: 1,
    scale: 1,
    y: 0,
    transition: {
      type: "spring" as const,
      stiffness: 480,
      damping: 34,
      mass: 0.78,
    },
  },
  exit: {
    opacity: 0,
    scale: 1.04,
    y: 8,
    transition: { duration: 0.2, ease: appleEase },
  },
};
