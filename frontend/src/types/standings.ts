/** A single row in the league standings table. */
export interface StandingsRow {
  /** 1-based rank within the league. */
  position: number;
  /** Identifier of the manager this row belongs to. */
  userId: number;
  /** Manager's display name. */
  name: string;
  /** Manager's initials, e.g. "AP". */
  initials: string;
  /** Name of the manager's squad. */
  squadName: string;
  /** Accumulated points across the season. */
  totalPoints: number;
  /** Points earned in the most recent matchday. */
  matchdayPoints: number;
  /** Change in position vs. the previous matchday (positive = moved up). */
  deltaPosition: number;
  /** Accent color associated with the row (CSS color string). */
  color: string;
  /** Whether this row represents the currently authenticated user. */
  isCurrentUser: boolean;
}
