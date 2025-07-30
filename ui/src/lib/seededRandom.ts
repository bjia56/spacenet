/**
 * Simple seedable pseudo-random number generator using a Linear Congruential Generator (LCG)
 * This provides deterministic random numbers for consistent visualizations
 */
export class SeededRandom {
  private seed: number;

  constructor(seed: number) {
    this.seed = seed % 2147483647;
    if (this.seed <= 0) this.seed += 2147483646;
  }

  /**
   * Generate next random number between 0 and 1
   */
  random(): number {
    this.seed = (this.seed * 16807) % 2147483647;
    return (this.seed - 1) / 2147483646;
  }

  /**
   * Generate random integer between min (inclusive) and max (exclusive)
   */
  randomInt(min: number, max: number): number {
    return Math.floor(this.random() * (max - min)) + min;
  }

  /**
   * Generate random float between min and max
   */
  randomFloat(min: number, max: number): number {
    return this.random() * (max - min) + min;
  }

  /**
   * Generate random boolean with given probability (0-1)
   */
  randomBoolean(probability: number = 0.5): boolean {
    return this.random() < probability;
  }

  /**
   * Pick random element from array
   */
  randomChoice<T>(array: T[]): T {
    return array[this.randomInt(0, array.length)];
  }
}