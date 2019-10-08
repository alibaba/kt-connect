import { connect } from 'react-redux'
import Connect from '../components/Connect'
import { actions } from '../modules'

const mapStateToProps = (state) => ({...state.connect})
export default connect(mapStateToProps, actions)(Connect)