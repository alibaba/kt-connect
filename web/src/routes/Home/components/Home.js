import React, { Component } from 'react';
import { Statistic, Card, Row, Col, Icon } from 'antd';
import './index.scss'

class Home extends Component {
  componentDidMount() {
    const { match } = this.props;
    this.props.fetchComponents(match.params.namespace);
  }

  render() {
    const { connect, exchange, mesh } = this.props;
    return (
      <div className="dashboard">
        <Row>
          <Col span={8}>
            <Card>
              <Statistic
                title="Connect"
                value={connect.length}
                precision={0}
                valueStyle={{ color: '#3f8600' }}
                prefix={<Icon type="desktop" />}
              />
            </Card>
          </Col>
          <Col span={8}>
            <Card>
              <Statistic
                title="Exchange"
                value={exchange.length}
                precision={0}
                valueStyle={{ color: '#3f8600' }}
                prefix={<Icon type="switcher" />}
              />
            </Card>
          </Col>
          <Col span={8}>
            <Card>
              <Statistic
                title="Mesh"
                value={mesh.length}
                precision={0}
                valueStyle={{ color: '#3f8600' }}
                prefix={<Icon type="deployment-unit" />}
              />
            </Card>
          </Col>
        </Row>
      </div>
    );
  }
}

export default Home;
