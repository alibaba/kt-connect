export const INCREMENT = 'INCREMENT'
export const INCREMENT_ASYNC = 'INCREMENT_ASYNC'

// Actions
export function increment(value = 1) {
  return {
    type: INCREMENT,
    payload: value
  }
}

export const incrementAsync = () => {
  return (dispatch, getState) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        dispatch({
          type: INCREMENT_ASYNC,
          payload: getState().setting.counter
        })
        resolve()
      }, 200)
    })
  }
}

export const actions = {
  increment,
  incrementAsync
}

// handlers
const ACTION_HANDLERS = {
  [INCREMENT]: (state, action) => { 
    return Object.assign({}, state, {counter: state.counter + action.payload}) 
  },
  [INCREMENT_ASYNC]: (state) => {
    return Object.assign({}, state, {counter: state.counter * 2}) 
  }
}

// reducer
const initialState = {
  counter: 0,
}
export default function reducer(state = initialState, action) {
  const handler = ACTION_HANDLERS[action.type]
  return handler ? handler(state, action) : state
}