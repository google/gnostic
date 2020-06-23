
3.0 
OpenAPI Petstore*
MIT21.0.06
 https://petstore.openapis.org/v1Development server"�
�
/pets�"�
petsList all pets*listPets2V
T
limitquery.How many items to return at one time (max 100)R
�integer�int32B�
L
J
unexpected error6
4
application/json 

#/components/schemas/Error�
200�
�
An paged array of petsA
?
x-next5
3
$A link to the next page of responsesB
	�string5
3
application/json

#/components/schemas/Pets2�
petsCreate a pet*
createPetsBh
L
J
unexpected error6
4
application/json 

#/components/schemas/Error
201

Null response
�
/pets/{petId}�"�
petsInfo for a specific pet*showPetById2=
;
petIdpathThe id of the pet to retrieve R
	�stringB�
L
J
unexpected error6
4
application/json 

#/components/schemas/Errorf
200_
]
$Expected response to a valid request5
3
application/json

#/components/schemas/Pets*�
�
]
PetV
T�id�name�E

id
�integer�int64

name
	�string

tag
	�string
3
Pets+
)�array�

#/components/schemas/Pet
U
ErrorL
J�code�message�6

code
�integer�int32

message
	�string