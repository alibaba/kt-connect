import { connect } from 'react-redux'
import Component from '../components/Service'
import { actions } from '../modules'

const mapStateToProps = (state) => ({...state.service})
export default connect(mapStateToProps, actions)(Component)