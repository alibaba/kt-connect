import React, { Component } from "react";
import OwnerReference from './OwnerReference';
import './index.scss';

export default class OwnerReferences extends Component {
  render() {
    const { ownerReferences, namespace, type, targetPorts } = this.props;
    return ownerReferences.map((ownerReference, index) => (
      <OwnerReference key={`ownerReference-${index}`} namespace={namespace} type={type} ownerReference={ownerReference} targetPorts={targetPorts} />
    ))
  }
}