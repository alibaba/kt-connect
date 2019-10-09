import React from 'react'
import { Provider } from 'react-redux'
import PropTypes from 'prop-types'
import { BrowserRouter as Router } from 'react-router-dom'
import { renderRoutes } from 'react-router-config'

class App extends React.Component {
  static propTypes = {
    store: PropTypes.object.isRequired,
    routes: PropTypes.array.isRequired,
  }

  shouldComponentUpdate() {
    return false
  }

  render() {
    return (
      <Provider store={this.props.store}>
        <Router>
          {renderRoutes(this.props.routes)}
        </Router>
      </Provider>
    )
  }
}

export default App
