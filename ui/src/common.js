const MONTHS = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
];

export class YearMonthBucket {
  constructor(id, totalCount) {
    this.id = id;
    const [year, month] = id.split('-');
    this.year = year;
    this.month = month;
    this.totalCount = totalCount;
    this.photos = [];
  }

  id() {
    return this.id;
  }

  totalCount() {
    return this.totalCount;
  }

  appendPhotos(photos) {
    this.photos = [...this.photos, ...photos];
  }

  get grouping() {
    return this.year;
  }

  get heading() {
    return `${this.year} ${MONTHS[this.month - 1]}`;
  }
}
