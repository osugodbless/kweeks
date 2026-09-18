// Built-in question banks — ready-made decks an instructor can pull into any
// quiz from the builder without retyping. Answers live only here (frontend) and
// are copied into a new quiz at insert time; nothing about the correct option is
// ever shipped to players by the app.

export interface BankQuestion {
  prompt: string;
  options: string[];
  correctIndex: number;
  /** Optional per-question time override; falls back to the quiz default. */
  durationMs?: number;
}

export interface QuestionBank {
  id: string;
  name: string;
  description: string;
  questions: BankQuestion[];
}

export const QUESTION_BANKS: QuestionBank[] = [
  {
    id: "bmoni-live-demo",
    name: "BMONI live demo",
    description: "11 questions for the live hackathon demo — host intros, BMONI facts, product and security.",
    questions: [
      {
        prompt: "What is the Kweeks host's name?",
        options: ["Samuel Momoh", "Godbless Ossu", "Samuel Adeyemi", "Momoh Samuel"],
        correctIndex: 0,
      },
      {
        prompt: "What is the Kweeks Co-Host's name?",
        options: ["Blessing Ossu", "Godwin Ossu", "Samuel Momoh", "Godbless Ossu"],
        correctIndex: 3,
      },
      {
        prompt: "When was BMONI founded?",
        options: ["2022", "2025", "2023", "2024"],
        correctIndex: 1,
      },
      {
        prompt: "Who is the founder of BMONI?",
        options: ["Ashwin Ravichandran", "Jørn Lyseggen", "Tayo Oviosu", "Guy-Bertrand Njoya"],
        correctIndex: 1,
      },
      {
        prompt: "Which prominent African tech leader is a strategic investor and talent partner for BMONI?",
        options: ["Aliko Dangote", "Tony Elumelu", "Patrice Motsepe", "Iyinoluwa Aboyeji"],
        correctIndex: 3,
      },
      {
        prompt:
          "Before BMONI finally launched to the public, the creators secretly worked on building it for how long?",
        options: ["2 months", "6 months", "1 year", "2 years"],
        correctIndex: 3,
      },
      {
        prompt: "Which of these is NOT something you can do with a BMONI multi-currency wallet?",
        options: [
          "Hold US Dollars",
          "Withdraw physical gold bars",
          "Swap Naira for Dollars",
          "Save money in stablecoins",
        ],
        correctIndex: 1,
      },
      {
        prompt:
          "BMONI lets users save money securely. What type of digital currency is used for their USD savings feature?",
        options: ["Bitcoin", "Dogecoin", "Ethereum", "Stablecoins"],
        correctIndex: 3,
      },
      {
        prompt: "What happens to your BMONI funds if your phone gets stolen or lost?",
        options: [
          "A thief can reset your password",
          "You lose your money forever",
          "Your funds stay safe because a thief doesn't have your face",
          "You have to buy a new phone from BMONI",
        ],
        correctIndex: 2,
      },
    ],
  },
];

/** Every bank question keyed as `<bankId>:<index>` for selection state. */
export function allBankKeys(): string[] {
  return QUESTION_BANKS.flatMap((bank) => bank.questions.map((_, i) => `${bank.id}:${i}`));
}
