import request from 'superagent'
import * as Util from '../../../utils'

export const GET_COMPONENTS = 'GET_COMPONENTS'

// Actions
export const fetchComponents = (namespace = 'default') => {
  return (dispatch) => {
    request
      .get(`/api/cluster/namespaces/${namespace}/components`)
      .then((res) => {
        if (res.body) {
          const group = Util.groupByComponent(res.body)
          dispatch({
            type: GET_COMPONENTS,
            payload: group,
          })
        }
      })
  }
}

export const actions = {
  fetchComponents,
}

// handlers
const ACTION_HANDLERS = {
  [GET_COMPONENTS]: (state, action) => {
    return Object.assign({}, state, { ...action.payload })
  }
}

// reducer
const initialState = {
  connect: [],
  exchange: [],
  mesh: []
}

export default function reducer(state = initialState, action) {
  const handler = ACTION_HANDLERS[action.type]
  return handler ? handler(state, action) : state
}