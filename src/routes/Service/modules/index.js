import request from 'superagent'

export const SET_SERVICE = 'SET_SERVICE';
export const SET_VIRTUAL_SERVICE = 'SET_VIRTUAL_SERVICE';
export const SET_ENDPOINT = 'SET_ENDPOINT';
export const SET_PODS = 'SET_PODS';
export const SET_DESTINATION_RULES = 'SET_DESTINATION_RULES';

// Actions
export const fetchService = (namespace, name) => {
  return (dispatch) => {
    request
      .get(`/api/cluster/namespaces/${namespace}/services/${name}`)
      .then((res) => {
        const service = res.body
        if (service) {
          dispatch({
            type: SET_SERVICE,
            payload: service,
          })
          dispatch(fetchEndpoint(namespace, name))
          dispatch(fetchPods(namespace, service.spec.selector))
        }
      })
  }
}

export const fetchVirtualService = (namespace, name) => {
  return (dispatch) => {
    request
      .get(`/api/cluster/namespaces/${namespace}/virtualservices/${name}`)
      .then((res) => {
        const data = res.body
        if (data) {
          dispatch({
            type: SET_VIRTUAL_SERVICE,
            payload: data,
          })
          dispatch(fetchDestinationRule(namespace, name))
        }
      })
  }
}

export const fetchDestinationRule = (namespace, name) => {
  return (dispatch) => {
    request
      .get(`/api/cluster/namespaces/${namespace}/destinationrules/${name}`)
      .then((res) => {
        const data = res.body
        if (data) {
          dispatch({
            type: SET_DESTINATION_RULES,
            payload: data,
          })
        }
      })
  }
}

export const fetchEndpoint = (namespace, name) => {
  return (dispatch) => {
    request
      .get(`/api/cluster/namespaces/${namespace}/endpoints/${name}`)
      .then((res) => {
        if (res.body) {
          dispatch({
            type: SET_ENDPOINT,
            payload: res.body,
          })
        }
      })
  }
}

export const fetchPods = (namespace, selector) => {
  console.log(JSON.stringify(selector));
  return (dispatch) => {
    request
      .get(`/api/cluster/namespaces/${namespace}/pods?selector=${JSON.stringify(selector)}`)
      .then((res) => {
        if (res.body) {
          dispatch({
            type: SET_PODS,
            payload: res.body,
          })
        }
      })
  }
}

export const actions = {
  fetchService,
  fetchVirtualService,
  fetchDestinationRule,
}

// handlers
const ACTION_HANDLERS = {
  [SET_SERVICE]: (state, action) => {
    return Object.assign({}, state, { service: action.payload })
  },
  [SET_ENDPOINT]: (state, action) => {
    return Object.assign({}, state, { endpoint: action.payload })
  },
  [SET_PODS]: (state, action) => {
    return Object.assign({}, state, { pods: action.payload })
  },
  [SET_VIRTUAL_SERVICE]: (state, action) => {
    return Object.assign({}, state, { virtualservice: action.payload })
  },
  [SET_DESTINATION_RULES]: (state, action) => {
    return Object.assign({}, state, { destinationrule: action.payload })
  }
}

// reducer
const initialState = {
  pods: [],
}

export default function reducer(state = initialState, action) {
  const handler = ACTION_HANDLERS[action.type]
  return handler ? handler(state, action) : state
}