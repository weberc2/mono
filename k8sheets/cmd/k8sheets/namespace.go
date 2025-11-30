package main

import v1 "k8s.io/api/core/v1"

var (
	namespacesSheet = sheet[v1.Namespace]{
		name:        sheetNamespaces,
		columnNames: namespaceColumnNames[:],
		values:      namespaceColumnValues[:],
	}

	namespaceColumnValues = [...]func(*v1.Namespace) string{
		namespaceColumnName: func(n *v1.Namespace) string {
			return n.Name
		},
		namespaceColumnStatus: func(n *v1.Namespace) string {
			return string(n.Status.Phase)
		},
		namespaceColumnCreated: func(n *v1.Namespace) string {
			return n.CreationTimestamp.Format(sheetsTimeFormat)
		},
	}

	namespaceColumnNames = [...]string{
		namespaceColumnName:    "Name",
		namespaceColumnStatus:  "Status",
		namespaceColumnCreated: "Created",
	}
)

const sheetNamespaces = "_Namespaces"

const (
	namespaceColumnName = iota
	namespaceColumnStatus
	namespaceColumnCreated
	namespaceColumnCOUNT // not a namespace column
)
