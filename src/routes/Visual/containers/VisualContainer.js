import { connect } from 'react-redux'
import Visual from '../components/Visual'
import { actions } from '../modules'

const mapStateToProps = (state) => ({...state.visual})
export default connect(mapStateToProps, actions)(Visual)