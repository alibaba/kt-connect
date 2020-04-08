import { connect } from 'react-redux'
import Setting from '../components/Setting'
import { increment, incrementAsync } from '../modules'

const mapDispatchToProps = {
  increment: () => increment(1),
  incrementAsync
}

const mapStateToProps = (state) => {
  return ({
    counter: state.setting.counter
  })
}

export default connect(mapStateToProps, mapDispatchToProps)(Setting)