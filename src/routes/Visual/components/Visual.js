import React, { Component } from 'react';
import * as R from 'ramda';
import { Modal, Empty } from 'antd';
import Graph from './Graph';
import Endpoint from './Endpoint/Endpoint';

import './visual.scss'

class Visual extends Component {

  constructor(props) {
    super(props)
    const { match } = props;
    this.state = {
      namespace: match.params.namespace,
      visible: false,
      endpoint: null,
    }
  }

  componentDidMount() {
    const { match } = this.props;
    const { namespace = 'default' } = match.params
    this.props.fetchComponents(namespace);
    this.props.fetchEndpoints(namespace)
  }

  render() {
    const { exchange, mesh, endpoints } = this.props;

    const group = R.groupBy((item) => {
      return item.status.podIP
    })

    const exchangeGroupByPodIP = group(exchange);
    const meshGroupByPodIP = group(mesh);
    const exchangeIPList = Object.keys(exchangeGroupByPodIP);
    const meshIPList = Object.keys(meshGroupByPodIP);

    const sortByExchange = R.sortBy((item) => {
      return item.proxyToLocal ? 0 : 1;
    });

    const formatEndpoints = endpoints
      .filter(item => {
        return item.subsets && item.subsets.length > 0
      }).map(item => {
        item.subsets.map((subset) => {
          const addresses = (subset.addresses || []).map((address) => {
            if (exchangeIPList.includes(address.ip)) {
              address.proxyStatus = true
              address.componentType = 'exchange'
              address.component = exchangeGroupByPodIP[address.ip]
            } else if (meshIPList.includes(address.ip)) {
              address.proxyStatus = true;
              address.componentType = 'mesh'
              address.component = meshGroupByPodIP[address.ip]
            } else {
              address.proxyStatus = false;
            }
            return address;
          })
          subset.addresses = addresses
          return subset
        })

        const proxyToLocal = item.subsets.find((subset) => {
          return subset.addresses && subset.addresses.find(address => {
            return address.proxyStatus
          })
        });

        return Object.assign(item, { proxyToLocal: proxyToLocal, proxyStatus: proxyToLocal !== undefined });
      })

    const allEndpoints = sortByExchange(formatEndpoints);

    return (
      <div>
        {allEndpoints.length > 0 &&
          <div>
            {/* <Row>
              <Col span={4}>Service</Col>
              <Col span={8}>Address</Col>
              <Col span={7}>Endpoints</Col>
              <Col span={3}>Status</Col>
              <Col span={2} style={{textAlign: 'center'}}>Operation</Col>
            </Row> */}
            {
              allEndpoints.map((endpoint, index) => (<Endpoint key={`endpoint-list-${index}`} namespace={this.state.namespace} onOpenSideMenuClick={this.props.onOpenSideMenuClick} endpoint={endpoint} onClick={() => {
                this.setState({ visible: true, endpoint: endpoint })
              }} />))
            }
          </div>
        }
        {
          allEndpoints.length === 0 && <Empty />
        }
        <Modal
          title="Topo View"
          visible={this.state.visible}
          onOk={() => { this.setState({ visible: false }) }}
          onCancel={() => { this.setState({ visible: false }) }}
        >
          <Graph item={this.state.endpoint} x={1} {...this.props} />
        </Modal>
      </div>
    );
  }
}

export default Visual;
