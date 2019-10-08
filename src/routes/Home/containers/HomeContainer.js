import { connect } from 'react-redux'
import Home from '../components/Home'
import { actions } from '../modules'

const mapStateToProps = (state) => ({...state.home})
export default connect(mapStateToProps, actions)(Home)