import React, { Component } from 'react';
import { Row, Col, Icon, Tag, Modal, Button } from 'antd';
import { Link } from 'react-router-dom'
import Countdown from 'react-countdown-now';
import Menus from '../Menus';
import './index.scss';

export default class Endpoint extends Component {
  render() {
    const { endpoint, onClick, onOpenSideMenuClick } = this.props;
    const addresses = endpoint.subsets.flatMap((subset) => {
      return subset.addresses;
    });

    const ports = endpoint.subsets.flatMap((subset) => subset.ports)
    const url = `${endpoint.metadata.name}.${endpoint.metadata.namespace}.svc.cluster.local`;

    return (<Row className={`endpoint-list-item ${endpoint.proxyStatus && 'endpoint-list-item-connected'}`}>
      <Col span={4} className="basic">
        <Link to={`/${this.props.namespace}/visual/${endpoint.metadata.name}`}><Icon type="global" size="large" />&nbsp;{endpoint.metadata.name}</Link>
      </Col>
      <Col span={8}>
        {ports.map((port, index) => {
          return (<Button key={`endpoint-btn-${index}`} type="link" onClick={() => {
            Modal.success({
              title: <div>New Window is opening, after <Countdown date={Date.now() + 3000} renderer={(count)=>{
                return <span>{count.seconds}...</span>
              }}/></div>,
              content: <div>
                <p>If fails to load page. you can use `ktctl connect` connect the cluster from localhost</p>
                <p className="command-line">ktctl connect</p>
              </div>
            });
            window.setTimeout(() => {
              window.open(`http://${url}:${port.port}`, '_blank');
            }, 3000);
          }}>
            <Icon type="link" />{`http://${url}:${port.port}`}
          </Button>)
        })}
      </Col>
      <Col span={7}>
        {
          addresses.map((address) => {
            return (
              <div
                className={`address address-icon address-${address.proxyStatus && 'proxy'}`}
                key={`endpoint-${address.ip}`}
                onClick={() => {
                  const type = address.proxyStatus ? 'component' : 'pod';
                  const Menu = Menus({ type, component: address.component ? address.component[0] : address.targetRef });
                  onOpenSideMenuClick({
                    content: Menu
                  });
                }}
              >
                {address.proxyStatus ? <Icon type='cloud-download' /> : <Icon type="cloud-server" />}
              </div>
            )
          })
        }
      </Col>
      <Col span={3}>
        {endpoint.proxyStatus ?
          <Tag color="#87d068">Redirected</Tag> :
          <Tag color="#108ee9">In house</Tag>
        }
      </Col>
      <Col span={2} style={{textAlign: 'center'}}>
        <Button type="default" shape="circle" icon="eye" onClick={onClick} size={'small'}/>&nbsp;
        {/* <Button type="default" shape="circle" icon="setting" onClick={()=>{console.log('redirect to setting')}} size={'small'}/> */}
      </Col>
    </Row>)
  }

}