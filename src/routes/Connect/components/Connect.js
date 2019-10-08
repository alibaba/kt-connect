import React, { Component } from 'react';
import Point from '../../../components/Point';
import NotFound from '../../../components/MessagePage/NotFound';

import './Connect.scss';

class Home extends Component {

  constructor(props) {
    super(props)
    const { match } = props;
    this.state = {
      namespace: match.params.namespace,
    }
  }

  componentDidMount() {
    const {match} = this.props;
    this.props.fetchComponents(match.params.namespace);
  }

  render() {
    const { connect } = this.props;
    return (
      <div>
        {connect.map((pod) => (<Point title={pod.metadata.name} />))}
        {
          connect.length === 0 &&
          <NotFound 
            icon="message"
            generalMessage="You can start a new connect from your terminal" 
            copyableMessge={`kubectl -namespace ${this.state.namespace} connect`}
          />
        }
      </div>
    );
  }
}

export default Home;
