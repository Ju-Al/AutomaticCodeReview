import formatDistance from "date-fns/formatDistance";
import parse from "date-fns/parse";
import format from "date-fns/format";

export function formatServerDate(dateString) {
  return `${formatDistance(new Date(dateString), new Date())} ago`;
}

export function monthDayYear(dateString) {
  return format(new Date(dateString), "MMM dd, uuuu");
}
