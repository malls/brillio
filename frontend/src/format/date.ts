// Listed-date formatting (BRANDING.md section 7.1). The API sends a calendar date, YYYY-MM-DD.
// It is read by parts and never turned into a Date: new Date('2026-08-29') is UTC midnight,
// which is Aug 28 in US time zones. Month names come from a fixed table, not the locale.

const MONTHS = [
  'Jan',
  'Feb',
  'Mar',
  'Apr',
  'May',
  'Jun',
  'Jul',
  'Aug',
  'Sep',
  'Oct',
  'Nov',
  'Dec',
] as const;

const ISO_DATE = /^(\d{4})-(\d{2})-(\d{2})$/;

// 'Listed Aug 29', or 'Listed Aug 29, 2025' outside currentYear. currentYear defaults to the
// local clock's year. A value that is not a valid YYYY-MM-DD is shown raw, so bad data stays
// visible.
export function formatListedDate(
  isoDate: string,
  currentYear: number = new Date().getFullYear(),
): string {
  const match = ISO_DATE.exec(isoDate);
  const year = Number(match?.[1]);
  const month = MONTHS[Number(match?.[2]) - 1];
  const day = Number(match?.[3]);
  if (month === undefined || !(day >= 1 && day <= 31)) {
    return `Listed ${isoDate}`;
  }
  const monthDay = `${month} ${day}`;
  return year === currentYear ? `Listed ${monthDay}` : `Listed ${monthDay}, ${year}`;
}
