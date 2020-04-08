import React, { Component } from 'react';
import { Icon } from 'antd';
import './index.scss';

class Point extends Component {
  render() {
    const { title, subTitle, className, icon } = this.props
    return (
      <div>
        <div className={`endpoint ${className}`}>
          <div className="cycle-wrapper" onClick={this.props.onClick}>
            <div className={`${icon ? 'icon' : 'default'}`}>
              {icon && <Icon type={icon} />}
            </div>
          </div>
          <div className="podName text">{title}</div>
          <div className="remoteIP text">{subTitle}</div>
        </div>

      </div>
    )
  }
}

export default Point