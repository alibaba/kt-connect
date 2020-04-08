import { connect } from 'react-redux'
import Component from '../components/Template'
import { actions } from '../modules'

// 1. change file name
const mapStateToProps = (state) => ({...state.home})
export default connect(mapStateToProps, actions)(Component)