package main

import (
	"sort"
	"strings"

	v1 "k8s.io/api/core/v1"
)

var (
	nodeSheet = sheet[v1.Node]{
		name:        sheetNodes,
		columnNames: nodeColumnNames[:],
		values:      nodeColumnValues[:],
	}

	nodeColumnNames = [...]string{
		nodeColumnName:    "Name",
		nodeColumnStatus:  "Status",
		nodeColumnRoles:   "Roles",
		nodeColumnCreated: "Created",
		nodeColumnVersion: "Version",
	}

	nodeColumnValues = [...]func(node *v1.Node) string{
		nodeColumnName: func(node *v1.Node) string { return node.Name },
		nodeColumnStatus: func(node *v1.Node) string {
			status := make([]string, 0, len(node.Status.Conditions))
			for i := range node.Status.Conditions {
				cond := &node.Status.Conditions[i]
				if cond.Status == v1.ConditionTrue {
					status = append(status, string(cond.Type))
				}
			}
			sort.Strings(status)
			return strings.Join(status, ",")
		},
		nodeColumnRoles: func(node *v1.Node) string {
			var role []string
			for key, value := range node.Labels {
				const prefix = "node-role.kubernetes.io/"
				if strings.HasPrefix(key, prefix) && value == "true" {
					role = append(role, key[len(prefix):])
				}
			}
			sort.Strings(role)
			return strings.Join(role, ",")
		},
		nodeColumnCreated: func(node *v1.Node) string {
			return node.CreationTimestamp.Time.Format(sheetsTimeFormat)
		},
		nodeColumnVersion: func(node *v1.Node) string {
			return node.Status.NodeInfo.KubeletVersion
		},
	}
)

const (
	nodeColumnName = iota
	nodeColumnStatus
	nodeColumnRoles
	nodeColumnCreated
	nodeColumnVersion
	nodeColumnCOUNT // not a nodeColumn
)
