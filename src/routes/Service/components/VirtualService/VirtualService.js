import React, { Component } from "react";
import { Card, Table, Tag, Button, Icon } from 'antd';
import CreateHttpRoute from './CreateHttpRoute';

export const MatchConig = (props) => {
  const { config } = props;
  if (config.uri) {
    return (<Tag>{config.uri.prefix}</Tag>)
  }
  return <Tag>Unknown Match Config</Tag>
}

export default class VirtualService extends Component {

  constructor(props) {
    super(props);
    this.state = {
      addDestinationVisible: false,
      current: 0,
    }
  }

  render() {
    const { virtualservice } = this.props;
    if (!virtualservice) {
      return null;
    }
    return (
      <div className="virtualservice-container">
        <CreateHttpRoute
          {...this.props}
          title="Add Route"
          visible={this.state.addDestinationVisible}
          onOk={() => { this.setState({ addDestinationVisible: false }) }}
          onCancel={() => { this.setState({ addDestinationVisible: false }) }}
        />
        <Card
          title={'Networking'}
          // extra={
          //   <Button onClick={() => {
          //     this.setState({
          //       addDestinationVisible: true,
          //     });
          //   }}>
          //     <Icon type="plus" />Add Destination
          //   </Button>
          // }
        >
          <Table dataSource={virtualservice.spec.http} pagination={false} columns={[
            {
              title: 'Route',
              key: 'Route',
              render: (httpRoute) => {
                const { match } = httpRoute;
                return !match ? <Tag>Default</Tag> :
                  match.map((config, index) => {
                    return <MatchConig key={`match-config-${index}`} config={config} />
                  })
              }
            }, {
              title: 'Destination',
              key: 'Destination',
              render: (httpRoute) => {
                const { route } = httpRoute;
                return route.map((config, index) => {
                  const destination = config.destination;
                  return (<Tag key={`route-config-${index}`}>{destination.subset}</Tag>)
                })
              }
            },
            {
              title: '',
              key: 'Ops',
              render: (httpRoute) => {
                return [
                  {
                    name: 'Edit',
                    icon: 'edit',
                    onClick: () => { console.log('add version') }
                  },
                  {
                    name: 'Remove',
                    icon: 'delete',
                    onClick: () => { console.log('add version') }
                  }
                ].map((btn, index) => {
                  return <Button key={`httpRoute-btn-${index}`} shape="circle"><Icon type={btn.icon} /></Button>
                })
              }
            }
          ]} />
        </Card>
      </div>
    );
  }
}