import os
import sys
import csv
import time
import numpy as np

# Create a vectorized function for calculating distances
import math
def sqrDiff(x1, x2):
	return (x2-x1)*(x2-x1)
vectSqrDiff = np.vectorize(sqrDiff)
def distFunc(v1, v2):
	return math.sqrt(np.sum(vectSqrDiff(v1, v2)))

# Open and import the original dataset
dataLabels = []
dataMatrixL = []
labelsRef = []
with open('../webkb.csv', 'rb') as csvfile:
	reader = csv.reader(csvfile)
	rowIndx = -1
	for row in reader:
		if rowIndx >= 0:
			dataMatrixL.append(np.zeros((len(row) - 1,), dtype=np.float64))
			for i in range(0, len(row)):
				if i < (len(row) - 1):
					dataMatrixL[rowIndx][i] = float(row[i])
				else:
					dataLabels.append(int(row[i]))
					if int(row[i]) not in labelsRef:
						labelsRef.append(int(row[i]))
		rowIndx = rowIndx + 1

# Open and import the resulting cendtroids (and time) file results
# from the command-line value
dataMatrixC = []
timeVal = ''
with open(sys.argv[1], 'rb') as csvfile:
	reader = csv.reader(csvfile)
	rowIndx = -1
	for row in reader:
		if len(row) == 1:
			timeVal = row[0]
		else:
			dataMatrixC.append(np.zeros((len(row),), dtype=np.float64))
			for i in range(0, len(row)):
				dataMatrixC[rowIndx][i] = float(row[i])	
		rowIndx = rowIndx + 1

# Determine the distances between each point and the nearest centriod
distVals = np.zeros((len(dataMatrixL),), dtype=np.float64)
distIndx = np.zeros((len(dataMatrixC),), dtype=np.int32)
bestCentroid = np.zeros((len(dataMatrixL),), dtype=np.int32)
confusionMatrix = np.zeros((len(dataMatrixC), len(labelsRef)), dtype=np.int32)
for i in range(0, len(dataMatrixL)):
	minDist = distFunc(dataMatrixC[0], dataMatrixL[i])
	minIndx = 0;
	for j in range(1, len(dataMatrixC)):
		dist = distFunc(dataMatrixC[j], dataMatrixL[i])
		if dist < minDist:
			minDist = dist
			minIndx = j
	distVals[i] = minDist
	distIndx[minIndx] = distIndx[minIndx] + 1
	bestCentroid[i] = minIndx

	# Continue to determine the proper label for the centroid
	lIndx = 0
	for item in labelsRef:
		if item == dataLabels[i]:
			break
		lIndx = lIndx + 1
	confusionMatrix[minIndx][lIndx] = confusionMatrix[minIndx][lIndx] + 1

# Create vectors of "guessed" labels
guessedLabels = np.zeros((len(dataMatrixL),), dtype=np.int32)
for i in range(0, len(dataMatrixL)):
	guessedLabels[i] = labelsRef[np.argmax(confusionMatrix[bestCentroid[i]])]

# Calculate the purity
maxRows = np.zeros((len(dataMatrixC)), dtype=np.float64)
for i in range(0, len(dataMatrixC)):
	maxRows[i] = np.max(confusionMatrix[:][i])
purity = np.sum(maxRows)/len(dataMatrixL)

# Write guessed labels to file
os.remove(sys.argv[1])
File = open(sys.argv[1], 'w')
for j in range(0, len(guessedLabels)):
	if j < (len(guessedLabels) - 1):
		File.write(str(guessedLabels[j]) + ",")
	else:
		File.write(str(guessedLabels[j]))
File.write('\n')
File.write(str(timeVal))
File.write('\n')
File.write(str(purity))
File.close()