import request from 'superagent'
import * as Util from '../../../utils'

export const GET_COMPONENTS = 'GET_COMPONENTS'
export const GET_ENDPOINTS = 'GET_ENDPOINTS'

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

export const fetchEndpoints = (namespace = 'default') => {
  return (dispatch) => {
    request
      .get(`/api/cluster/namespaces/${namespace}/endpoints`)
      .then((res) => {
        if (res.body) {
          dispatch({
            type: GET_ENDPOINTS,
            payload: res.body,
          })
        }
      })
  }
}

export const actions = {
  fetchComponents,
  fetchEndpoints,
}

// handlers
const ACTION_HANDLERS = {
  [GET_COMPONENTS]: (state, action) => {
    return Object.assign({}, state, { ...action.payload })
  },
  [GET_ENDPOINTS]: (state, action) => {
    return Object.assign({}, state, { endpoints: action.payload })
  }
}

// reducer
const initialState = {
  endpoints: [],
  connect: [],
  exchange: [],
  mesh: []
}

export default function reducer(state = initialState, action) {
  const handler = ACTION_HANDLERS[action.type]
  return handler ? handler(state, action) : state
}