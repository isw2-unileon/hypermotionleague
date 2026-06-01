import type { StandingsRow } from "@/types/standings";

function cycleColor(index: number): string {
  switch (index % 4) {
    case 0:
      return "var(--pos-mid)";
    case 1:
      return "var(--pos-def)";
    case 2:
      return "var(--pos-fwd)";
    default:
      return "var(--pos-gk)";
  }
}

function initialsFrom(name: string): string {
  return name
    .split(" ")
    .map((part) => part.charAt(0).toUpperCase())
    .join("");
}

const rows = [
  { position: 1, name: "Carlos Mendoza", squadName: "Los Pichichis", totalPoints: 2104, matchdayPoints: 92, deltaPosition: 1 },
  { position: 2, name: "Lucía Ramírez", squadName: "11 de Bilbao", totalPoints: 2087, matchdayPoints: 78, deltaPosition: 1 },
  { position: 3, name: "Andrés Peral", squadName: "Mis panas", totalPoints: 2042, matchdayPoints: 87, deltaPosition: 2 },
  { position: 4, name: "Iván Castaño", squadName: "FC Resaca", totalPoints: 2018, matchdayPoints: 64, deltaPosition: -2 },
  { position: 5, name: "Marina Soto", squadName: "Sotomayor", totalPoints: 1995, matchdayPoints: 71, deltaPosition: 0 },
  { position: 6, name: "Pablo Egea", squadName: "Egea Athletic", totalPoints: 1948, matchdayPoints: 56, deltaPosition: -1 },
  { position: 7, name: "Rocío Vidal", squadName: "Vidal United", totalPoints: 1903, matchdayPoints: 82, deltaPosition: 3 },
  { position: 8, name: "Diego Hernán", squadName: "DH Galácticos", totalPoints: 1872, matchdayPoints: 49, deltaPosition: -2 },
  { position: 9, name: "Sara Bilbao", squadName: "Olé Olé Olé", totalPoints: 1841, matchdayPoints: 67, deltaPosition: 0 },
  { position: 10, name: "Jorge Vázquez", squadName: "VAR de Caracas", totalPoints: 1799, matchdayPoints: 41, deltaPosition: -1 },
];

export const mockStandings: StandingsRow[] = rows.map((row, index) => ({
  ...row,
  userId: index + 1,
  initials: initialsFrom(row.name),
  color: cycleColor(index),
  isCurrentUser: false,
}));
