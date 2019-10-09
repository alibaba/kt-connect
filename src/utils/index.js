import * as R from 'ramda'

export const groupByComponent = R.groupBy((item) => {
  return item.metadata.labels['kt-component']
});