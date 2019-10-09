import React, { Component } from 'react';
import { Icon } from 'antd';

import './index.scss';

export default class NotFound extends Component {
  render() {
    const { icon, generalMessage, copyableMessge } = this.props;
    return (<div className="panel-404">
      <Icon type={icon} style={{ fontSize: '60px' }} />
      <p>{generalMessage}</p>
      <p className="code">{copyableMessge}</p>
    </div>)
  }
}