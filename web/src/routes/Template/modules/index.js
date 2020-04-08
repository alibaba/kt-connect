export const ACTION_TYPE = 'ACTION_TYPE'

// Actions
export const doAction = () => {
  return (dispatch) => {
    dispatch({
      type: ACTION_TYPE,
      payload: {data: 1},
    })
  }
}

export const actions = {
  doAction,
}

// handlers
const ACTION_HANDLERS = {
  [ACTION_TYPE]: (state, action) => {
    return Object.assign({}, state, { ...action.payload })
  }
}

// reducer
const initialState = {
}

export default function reducer(state = initialState, action) {
  const handler = ACTION_HANDLERS[action.type]
  return handler ? handler(state, action) : state
}